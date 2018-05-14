<?php
use Workerman\Worker;
use think\App;
use think\Db;
use PhpAmqpLib\Connection\AMQPStreamConnection;
use PhpAmqpLib\Message\AMQPMessage;

require_once ROOT_PATH . 'service/WalletSync.php';
require_once ROOT_PATH . 'service/WalletSyncAll.php';
require_once ROOT_PATH . 'service/CommonSync.php';

//日志
Worker::$stdoutFile = ROOT_PATH."service/workerman.log";

$consumer = new Worker();

// 慢任务，消费者的进程数可以开多一些
$consumer->count = 1;

//进程启动
$consumer->onWorkerStart = function($consumer) {
    App::initCommon();

    $connection = new AMQPStreamConnection(conf('mq_host'), conf('mq_port'), conf('mq_user'), conf('mq_pass'), conf('mq_vhost'));

    //日志
    restore_error_handler();
    set_error_handler('LogService::appError');

    //从商城mq配置上拉取数据
    $query = conf('mq_query_js');
    $exchange = conf('mq_exchange_js');

    //通过链接获得一个新通道.
    $channel = $connection->channel();



    $channel->queue_declare($query, false, true, false, false);

    //绑定消息队列和交换机
    $channel->queue_bind($query, $exchange);

    $channel->basic_consume($query, 'consumer', false, false, false, false, 'processMessage');

    //注册关闭进程调用
    register_shutdown_function('shutdown', $channel, $connection);

    /**
     * 循环等待消息
     * Loop as long as the channel has callbacks registered
     */
    while(count($channel->callbacks)) {
        $channel->wait();
    }
};

//进程关闭
$consumer->onWorkerStop = function($consumer) {
    restore_error_handler();
};

/**
 * 消息回调函数
 * @param \PhpAmqpLib\Message\AMQPMessage $message
 */
function processMessage($message)
{
    //获取消息内容
    $message_body = $message->body;
    if(empty($message_body)) {
        //消息体为空，不处理
        $message->delivery_info['channel']->basic_ack($message->delivery_info['delivery_tag']);
        return true;
    } else {
        //解码
        $message_data = json_decode(base64_decode($message_body), true);


        //消息记录
        Db::startTrans();
        $rs = Db::name('mq_msg_get')->where(['msgid' => $message_data['MsgId']])->find();
        if($rs) {
            //已处理过
            Db::commit();
            echo 'success';
        } else {

            //调用业务代码
            $res = call_user_func([$message_data['Service'], $message_data['Method']], $message_data['Args']);

            var_dump(json_encode($res));

            if(!$res['status']) {
                //处理失败
                Db::rollback();
                echo $res;
                return false; //必须return false 否则就被消费了
            } else {

                if ($message_data['type'] == "rpc"){

                    $sendData = base64_encode(json_encode($res['data']));

                    $msg = new AMQPMessage(
                        $sendData,
                        array('correlation_id' => $message->get('correlation_id'))
                    );
                    $message->delivery_info['channel']->basic_publish(
                        $msg, '', $message->get('reply_to'));
                }

                //处理成功
                $msg_data = array(
                    'msgid' => $message_data['MsgId'],
                    'data' => $message_body,
                );
                Db::name('mq_msg_get')->insert($msg_data);
                Db::commit();
                echo 'success';
            }
        }



        //确认消息处理成功
        $message->delivery_info['channel']->basic_ack($message->delivery_info['delivery_tag']);
    }

    // Send a message with the string "quit" to cancel the consumer.
    if($message->body === 'quit') {
        $message->delivery_info['channel']->basic_cancel($message->delivery_info['consumer_tag']);
    }
}

/**
 * 关闭连接回调
 * @param \PhpAmqpLib\Channel\AMQPChannel $channel
 * @param \PhpAmqpLib\Connection\AbstractConnection $connection
 */
function shutdown($channel, $connection)
{
    $channel->close();
    $connection->close();
}

//日志
class LogService
{
    // 日志级别 从上到下，由低到高
    const EMERG = 'EMERG';  // 严重错误: 导致系统崩溃无法使用
    const ALERT = 'ALERT';  // 警戒性错误: 必须被立即修改的错误
    const CRIT = 'CRIT';  // 临界值错误: 超过临界值的错误，例如一天24小时，而输入的是25小时这样
    const ERR = 'ERR';  // 一般错误: 一般性错误
    const WARN = 'WARN';  // 警告性错误: 需要发出警告的错误
    const NOTICE = 'NOTIC';  // 通知: 程序可以运行但是还不够完美的错误
    const INFO = 'INFO';  // 信息: 程序输出信息
    const DEBUG = 'DEBUG';  // 调试: 调试信息
    const SQL = 'SQL';  // SQL：SQL语句 注意只在调试模式开启时有效

    public static function record($message, $level = self::INFO)
    {
        $mem = intval(memory_get_usage() / 1024 / 1024);
        $msg = date('Ymd H:i:s', time()).' '.$level.' ['.str_pad($mem, 4, " ", STR_PAD_LEFT).'M] '.$message;
        echo $msg."\n";
    }

    static public function appError($errno, $errstr, $errfile, $errline)
    {
        if($errno == E_ERROR) {

            $errorStr = "$errstr ".$errfile." 第 $errline 行.";
            self::record("[$errno] $errstr ".$errfile." 第 $errline 行.", self::ERR);
        }

    }

    function heat()
    {

        //调试模式下输出错误信息
        $trace = debug_backtrace();
        //$e['message'] = $error;
        $e['file'] = $trace[0]['file'];
        $e['class'] = $trace[0]['class'];
        $e['function'] = $trace[0]['function'];
        $e['line'] = $trace[0]['line'];
        $traceInfo = '';
        $time = date('y-m-d H:i:m');
        foreach($trace as $t) {
            $traceInfo .= '['.$time.'] ';


            if(isset($t['file'])) {
                $traceInfo .= $t['file'];
            }
            if(isset($t['line'])) {
                $traceInfo .= ' ('.$t['line'].') ';
            }
            if(isset($t['class'])) {
                $traceInfo .= $t['class'];
            }


            $traceInfo .= $t['function'].'(';
            //$traceInfo .= implode(', ', $t['args']);
            $traceInfo .= ')'."\n";
        }
        echo $traceInfo;
        //$e['trace'] = $traceInfo;

    }

    function clear()
    {
        $cmd = "echo -ne \"\033[2J\n\"";
        $a = exec($cmd);
        print "$a"."\n";
    }

    function color_b($string, $line = 0)
    {
        $cmd = "printf \"\033[".$line.";0H \033[01;40;32m".$string."\033[0m\n\"";
        $a = exec($cmd);
        print "$a"."\n";
    }
}
