<?php
namespace app\common\logic\mq;
use PhpAmqpLib\Message\RpcClient;  //rpc消息
use think\Db;
use think\Model;
use PhpAmqpLib\Message\MsgConfirmPublisher;  //mq消息
use PhpAmqpLib\Connection\AMQPStreamConnection;
use PhpAmqpLib\Message\AMQPMessage;


/**
 * 发送MQ
 */
class SendMQ extends Model
{
    protected $log_file;

    protected $connection;
    protected $channel;
    protected $callback_queue;
    protected $response;
    protected $corr_id;
    protected $init =false;
    protected $queue;
    protected $exchange;
    /*
     * 构造函数
     */
    public function __construct()
    {
        $log_dir = RUNTIME_PATH.'mqlog/mqsend/';
        if (!is_dir($log_dir))
        {
            mkdir($log_dir, 0755, true);
        }
        $this->log_file = $log_dir.date('Y_m_d').'.log';
    }
    /**
     * 记录来往参数
     * @param unknown $msg
     */
    public function record($msg)
    {
        file_put_contents($this->log_file, date('H:i:s')."\n".$msg."\n", FILE_APPEND);
    }
    
    /**
     * 2.0传商城的一些表数据与自定义数据
     * @param string $symbol （选传）不传时只同步自定义数据
     * @param string $mark  （必传）同步时间点标志
     * @param array $custom  （选传） 自定义数据，原样传到结算
     * @return boolean
     */
    public function SendTableInfo($symbol='',$mark='',$custom=[])
    {
        $this->record("触发同步，symbol:".$symbol.';mark:'.$mark.';custom'.serialize($custom));
        if (empty($mark)){
            return false;
        }
        //开关开启判断
        if(empty(CONF('mq_open_status'))){
            $this->record("未开启A");
            return false;  //未开启
        }
        //查询数据
        $map = [];
        $map['mark'] = $mark;
        $map['status'] = 1;
        $mq_config = Db::name('mq_config')->where($map)->field('status,table_uname,face_service,face_method,p_data,symbol')->find();
        if (empty($mq_config)){
            $this->record("未开启B");
            return false;  //未开启
        }
        if (empty($mq_config['face_service']) || empty($mq_config['face_method'])){
            $this->record("数据不全");
            return false;  //数据不全
        }
        $pData = [];
        //发送表中的数据
        if (!empty($symbol) && !empty($mq_config['symbol']) && !empty($mq_config['p_data']) && !empty($mq_config['table_uname'])){
            if (isSerialized($mq_config['p_data'])){
                $p_data = unserialize($mq_config['p_data']);
            }else{
                $p_data = '';
            }
            $rFields = '';
            $rField  = [];
            foreach ($p_data as $key=>$var){
                if ($var['opt']){ //是开启同步的字段
                    $rFields = $rFields.$key.',';
                    $rField[$key] = $var;
                }
            }
            $rFields = trim($rFields,',');
            $this->record("同步的数据库字段".$rFields);
            if (!empty($rFields)){
                //查询要发送的数据
                $post_info = Db::name($mq_config['table_uname'])->where([$mq_config['symbol']=>$symbol])->field($rFields)->find();
                if (!empty($post_info)){
                    foreach ($post_info as $key=>$var){
                        if (!empty($rField[$key]['js'])){
                            $pData[$rField[$key]['js']] = $var;
                        }else{
                            $pData[$key] = $var;
                        }
                    }
                }
            }
        }
        //发送自定义数据(自定义数据的优先级在表中数据之上，存在重复的，直接使用自定义的覆盖表中的)
        if (!empty($custom)){
            foreach ($custom as $key=>$var){
                $pData[$key] = $var;
            }
        }
        //是否有要同步的数据
        if (empty($pData)){
            $this->record("没有要同步的数据");
            return false;  //理论上不应该走到这
        }
        
        //查询结算的配置信息
        $scConfig = logic('mq/GetMqConfig')->getScConfig('js');
        if ($scConfig){
            $query = $scConfig['query_js'];
            $exchange = $scConfig['exchange_js'];
        }else {
            $this->record("没有配置结算mq信息");
            return false;
        }
        
        $msg_id = getMsgIds();   //随机数，编号（也可以使用getMsgId();这个带s的是结算提供的一个生成方法）
        $msg_data = [];
        $msg_data['MsgId'] = $msg_id;
        
        $msg_data['Service'] = $mq_config['face_service'];
        $msg_data['Method'] = $mq_config['face_method'];
        
        $msg_data['Args']['data'] = $pData;
        $this->record("发送的总数据".serialize($msg_data['Args']));
        //json序列化并且base64编码
        $msg_data = base64_encode(json_encode($msg_data));
        
        //构造保存数据
        $msg_save = ['msgid' => $msg_id, 'data' => $msg_data];
        
        Db::name('mq_msg_send')->insert($msg_save);
        
        $msg_publisher = new MsgConfirmPublisher();
        $msg_publisher->queue = $query;
        
        //这个while感觉应该加个保护机制，要不然容易死循环
        $result = '';
        $p_flg = 0;
        while(!$result && $p_flg<10){
            $p_flg ++;
            $result = $msg_publisher->publisher($msg_data, $exchange, $query);
        }
        if ($result){
            $this->record("发送成功");
            Db::name('mq_msg_send')->where('msgid', $msg_id)->update(['status'=>1]);
        }else{
            $this->record("发送失败");
            //不做更改，意味本条同步失败了，后期人工处理
        }
        return true;
    }


    public function rpcInit(){

        if (!$this->init){

            $jsConfig = logic('mq/GetMqConfig')->getScConfig('js');


            $this->queue = "rpc_qujs";
            $this->exchange =  "rpc_exjs";

            $scConfig = logic('mq/GetMqConfig')->getScConfig('sc');

            $this->scQueue = "rpc_qusc";
            
            $this->connection = new AMQPStreamConnection(conf('mq_host'), conf('mq_port'), conf('mq_user'), conf('mq_pass'), conf('mq_vhost'));
            
            $this->channel = $this->connection->channel();
            
            $this->channel->queue_bind($this->queue, $this->exchange);
            
            list($this->callback_queue, ,) = $this->channel->queue_declare(
                $this->scQueue, false, true, false, false);
            
            $this->channel->basic_consume(
                $this->callback_queue, '', false, false, false, false,
                array($this, 'rpcResponse'));

            $this->init= true;
        }


       }

       public function rpcSend($class,$data){

           $this->rpcInit();
           
           $this->response = null;
           $this->corr_id = uniqid();


           $sendData = $this->rpcInitData($class,$data);

           $msg = new AMQPMessage(
               $sendData,
               array('correlation_id' => $this->corr_id,
                   'reply_to' => $this->callback_queue)
           );


           $this->channel->basic_publish($msg, "",$this->queue);

           while(!$this->response) {
               $this->channel->wait();
           }
           return  json_decode(base64_decode($this->response), true);

       }


       public function rpcInitData($class,$data){
           $msg_id = getMsgIds();   //随机数，编号（也可以使用getMsgId();这个带s的是结算提供的一个生成方法）
           $msg_data = [];
           $msg_data['MsgId'] = $msg_id;

           $msg_data['Service'] = $class;
           $msg_data['Method'] ="run";
           $msg_data['type'] ="rpc";
           $msg_data['Args']['data'] = $data;


           $this->record("rpc--------------接收类：{$class} 发送的rpc总数据为：".serialize($msg_data['Args']));
           //json序列化并且base64编码
           $msg_data = base64_encode(json_encode($msg_data));

           //构造保存数据
           $msg_save = ['msgid' => $msg_id, 'data' => $msg_data];

           Db::name('mq_msg_send')->insert($msg_save);

           return $msg_data;
       }

    public function rpcResponse($rep) {

        try{
        $rec_id = $rep->get('correlation_id');

        }catch (\Exception $e){
            $rec_id = 0;
        }
        if($rec_id && $rec_id == $this->corr_id) {
            $this->response = $rep->body;
            $this->record("rpc--------------被消费 接收的rpc数据为".base64_decode( $rep->body));

            $rep->delivery_info['channel']->basic_ack(
                $rep->delivery_info['delivery_tag']);
        }else{
            $this->record("rpc--------------未被消费 接收的rpc数据为".base64_decode( $rep->body));

            return false;
        }

    }
    /**
     * 传购物订单给结算
     * 传入参数order_id或者order_no
     */
    /* public function SendOrderInfo($order_id='',$order_no='')
    {
        //order_id与order_no必需要保证一个不为空
        if (empty($order_id) && empty($order_no)){
            return false;
        }
        //开关开启判断
        $openStatus = logic('mq/GetMqConfig')->getOpenStatus('order');
        if (isset($openStatus['mq_open_status']) && isset($openStatus['mq_open_order'])){
            if (!$openStatus['mq_open_status'] || !$openStatus['mq_open_order']){
                return false;
            }
        }else {
            return false;
        }
        
        //查询结算的配置信息
        $scConfig = logic('mq/GetMqConfig')->getScConfig('sc');
        if ($scConfig){
            $query = $scConfig['query_js'];
            $exchange = $scConfig['exchange_js'];
        }else {
            return false;
        }
        //查询要同步的数据（此处只查询需要的信息，虽然增加了一定的出错风险，但能提高不少效率）
        $field = logic('mq/GetMqConfig')->getMqField('order');
        if ($field){
            //查询订单信息
            if (!empty($order_no)){
                $order_ninfo = Db::name('orders')->where(['order_no'=>$order_no])->field($field['fields'])->find();
            }elseif(!empty($order_id)){
                $order_ninfo = Db::name('orders')->where(['id'=>$order_id])->field($field['fields'])->find();
            }else{
                return false;
            }
            //如果查询不到信息，直接返回
            if (empty($order_ninfo)){
                return false;
            }
        }else{
            return false;
        }
        
        //要传输的数据(理论上此时查询出来的都要传输，根据配置的结算接收字段，组织传输的数据)
        $pData = [];
        foreach ($order_ninfo as $key=>$var){
            if (!empty($field['field'][$key]['js'])){
                $pData[$field['field'][$key]['js']] = $var;
            }else{
                $pData[$key] = $var;
            }
        }
        if (empty($pData)){
            return false;  //理论上走不到这里
        }
        
        $msg_id = getMsgIds();   //随机数，编号（也可以使用getMsgId();这个带s的是结算提供的一个生成方法）
        $msg_data = [];
        $msg_data['MsgId'] = $msg_id;
        
        //获取结算接收的接口配置
        $port = logic('mq/GetMqConfig')->getMqPort('order');
        if ($port){
            $msg_data['Service'] = $port['mq_service'];
            $msg_data['Method'] = $port['mq_method'];
        }else {
            return false;
        }
        
        $msg_data['Args']['data'] = $pData;
        
        //json序列化并且base64编码
        $msg_data = base64_encode(json_encode($msg_data));
        
        //构造保存数据
        $msg_save = ['msgid' => $msg_id, 'data' => $msg_data];
        
        Db::name('mq_msg_send')->insert($msg_save);
        
        $msg_publisher = new MsgConfirmPublisher();
        $msg_publisher->queue = $query;
        
        //这个while感觉应该加个保护机制，要不然容易死循环
        $result = '';
        $p_flg = 0;
        while(!$result && $p_flg<10){
            $p_flg ++;
            $result = $msg_publisher->publisher($msg_data, $exchange, $query);
        }
        if ($result){
            Db::name('mq_msg_send')->where('msgid', $msg_id)->update(['status'=>1]);
        }else{
            //不做更改，意味本条同步失败了，后期人工处理
        }
        return true;
    } */
    
    /**
     * 发送会员信息到结算
     * @param string $user_id
     * @param string $username
     * @return boolean
     */
    /* public function sendUserInfo($user_id='',$username='')
    {
        //$user_id与$username必需要保证一个不为空
        if (empty($user_id) && empty($username)){
            return false;
        }
        //开关开启判断
        $openStatus = logic('mq/GetMqConfig')->getOpenStatus('user');
        if (isset($openStatus['mq_open_status']) && isset($openStatus['mq_open_user'])){
            if (!$openStatus['mq_open_status'] || !$openStatus['mq_open_user']){
                return false;
            }
        }else {
            return false;
        }
        
        //查询结算的配置信息
        $scConfig = logic('mq/GetMqConfig')->getScConfig('sc');
        if ($scConfig){
            $query = $scConfig['query_js'];
            $exchange = $scConfig['exchange_js'];
        }else {
            return false;
        }
        //查询要同步的数据（此处只查询需要的信息，虽然增加了一定的出错风险，但能提高不少效率）
        $field = logic('mq/GetMqConfig')->getMqField('user');
        if ($field){
            //查询用户信息
            if (!empty($user_id)){
                $user_ninfo = Db::name('user')->where(['id'=>$user_id])->field($field['fields'])->find();
            }elseif(!empty($username)){
                $user_ninfo = Db::name('user')->where(['username'=>$username])->field($field['fields'])->find();
            }else{
                return false;
            }
            //如果查询不到信息，直接返回
            if (empty($user_ninfo)){
                return false;
            }
        }else{
            return false;
        }
        
        //要传输的数据(理论上此时查询出来的都要传输，根据配置的结算接收字段，组织传输的数据)
        $pData = [];
        foreach ($user_ninfo as $key=>$var){
            if (!empty($field['field'][$key]['js'])){
                $pData[$field['field'][$key]['js']] = $var;
            }else{
                $pData[$key] = $var;
            }
        }
        if (empty($pData)){
            return false;  //理论上走不到这里
        }
        
        $msg_id = getMsgIds();   //随机数，编号（也可以使用getMsgId();这个带s的是结算提供的一个生成方法）
        $msg_data = [];
        $msg_data['MsgId'] = $msg_id;
        
        //获取结算接收的接口配置
        $port = logic('mq/GetMqConfig')->getMqPort('user');
        if ($port){
            $msg_data['Service'] = $port['mq_service'];
            $msg_data['Method'] = $port['mq_method'];
        }else {
            return false;
        }
        
        $msg_data['Args']['data'] = $pData;
        
        //json序列化并且base64编码
        $msg_data = base64_encode(json_encode($msg_data));
        
        //构造保存数据
        $msg_save = ['msgid' => $msg_id, 'data' => $msg_data];
        
        Db::name('mq_msg_send')->insert($msg_save);
        
        $msg_publisher = new MsgConfirmPublisher();
        $msg_publisher->queue = $query;
        
        //这个while感觉应该加个保护机制，要不然容易死循环
        $result = '';
        $p_flg = 0;
        while(!$result && $p_flg<10){
            $p_flg ++;
            $result = $msg_publisher->publisher($msg_data, $exchange, $query);
        }
        if ($result){
            Db::name('mq_msg_send')->where('msgid', $msg_id)->update(['status'=>1]);
        }else{
            //不做更改，意味本条同步失败了，后期人工处理
        }
        return true;
    } */
    
    /**
     * 发送钱包金额加减信息到结算
     * @param string $username
     * @param string $number
     * @param string $money_type
     * @param string $prize_type
     * @return boolean
     */
    public function sendWalletInfo($username='',$number='',$money_type='',$prize_type='')
    {
        $this->record("触发自动同步钱包");
        if (empty($username) || empty($money_type) || empty($number)){
            return false;
        }
        //开关开启判断
        $openStatus = logic('mq/GetMqConfig')->getOpenStatus('wallet');
        if (isset($openStatus['mq_open_status']) && isset($openStatus['mq_open_wallet'])){
            if (!$openStatus['mq_open_status'] || !$openStatus['mq_open_wallet']){
                return false;
            }
        }else {
            return false;
        }
        
        //查询结算的配置信息
        $scConfig = logic('mq/GetMqConfig')->getScConfig('js');
        if ($scConfig){
            $query = $scConfig['query_js'];
            $exchange = $scConfig['exchange_js'];
        }else {
            return false;
        }
        //组织要传输的数据
        $pData = [];
        $pData['money_type'] = $money_type;
        $pData['prize_type'] = $prize_type;
        $pData['username'] = $username;
        if ($number > 0){
            $pData['money_edit'] = 'add';
        }else{
            $pData['money_edit'] = 'minus';
        }
        $pData['number'] = abs($number);
        
        $msg_id = getMsgIds();   //随机数，编号（也可以使用getMsgId();这个带s的是结算提供的一个生成方法）
        $msg_data = [];
        $msg_data['MsgId'] = $msg_id;
        
        //获取结算接收的接口配置
        $port = logic('mq/GetMqConfig')->getMqPort('wallet');
        if ($port){
            $msg_data['Service'] = $port['mq_service'];
            $msg_data['Method'] = $port['mq_method'];
        }else {
            return false;
        }
        
        $msg_data['Args']['data'] = $pData;
        $this->record("发送的总数据".serialize($msg_data['Args']));
        //json序列化并且base64编码
        $msg_data = base64_encode(json_encode($msg_data));
        
        //构造保存数据
        $msg_save = ['msgid' => $msg_id, 'data' => $msg_data];
        
        Db::name('mq_msg_send')->insert($msg_save);
        
        $msg_publisher = new MsgConfirmPublisher();
        $msg_publisher->queue = $query;
        
        //这个while感觉应该加个保护机制，要不然容易死循环
        $result = '';
        $p_flg = 0;
        while(!$result && $p_flg<10){
            $p_flg ++;
            $result = $msg_publisher->publisher($msg_data, $exchange, $query);
        }
        if ($result){
            Db::name('mq_msg_send')->where('msgid', $msg_id)->update(['status'=>1]);
            $this->record("发送成功");
        }else{
            $this->record("发送失败");
            //不做更改，意味本条同步失败了，后期人工处理
        }
        return true;
    }
    
    //测试数据(自己发自己收)
   public function sendCommonInfo()
    {
        //增加传递结算消息
        $query = 'mtgjys_qusc';
        $exchange = 'mtgjys_exsc';
        
        $post_data = [];
        //钱包单条测试
       /*  $post_data['username'] = '888888';
        $post_data['money_type'] = 'money';
        $post_data['prize_type'] = '测试数据';
        $post_data['money_edit'] = 'add';
        $post_data['number'] = 1.11; */
        //钱包多条测试
        $post_data['money_type'] = 'money';
        $post_data['p_data'][0]['username'] = '104353';
        $post_data['p_data'][0]['prize_type'] = '测试数据';
        $post_data['p_data'][0]['money_edit'] = 'add';
        $post_data['p_data'][0]['number'] = 2;
        $post_data['p_data'][1]['username'] = '104353';
        $post_data['p_data'][1]['prize_type'] = '测试数据';
        $post_data['p_data'][1]['money_edit'] = 'add';
        $post_data['p_data'][1]['number'] = 5;
        $post_data['p_data'][2]['username'] = '888888';
        $post_data['p_data'][2]['prize_type'] = '测试数据';
        $post_data['p_data'][2]['money_edit'] = 'add';
        $post_data['p_data'][2]['number'] = 5;
        $post_data['p_data'][3]['username'] = '888888';
        $post_data['p_data'][3]['prize_type'] = '测试数据';
        $post_data['p_data'][3]['money_edit'] = 'minus';
        $post_data['p_data'][3]['number'] = 2;
        
        $msg_id = getMsgIds();
        
        $msg_data = [];
        $msg_data['MsgId'] = $msg_id;
        
        //$msg_data['Service'] = 'WalletSync';  //单条钱包
        $msg_data['Service'] = 'WalletSyncAll';  //批量钱包
        $msg_data['Method'] = 'run';
        
        $msg_data['Args']['data'] = $post_data;
        
        //json序列化并且base64编码
        $msg_data = base64_encode(json_encode($msg_data));
        
        //构造保存数据
        $msg_save = ['msgid' => $msg_id, 'data' => $msg_data];
        
        Db::name('mq_msg_send')->insert($msg_save);
        
        $msg_publisher = new MsgConfirmPublisher();
        $msg_publisher->queue = $query;
        
        $result = '';
        while(!$result){
            $result = $msg_publisher->publisher($msg_data, $exchange, $query);
        }
        Db::name('mq_msg_send')->where('msgid', $msg_id)->update(['status'=>1]);
        return true;
    }
    
    //测试数据(自己发自己收)
    /* public function sendCeshi()
    {
        //增加传递结算消息
        $query = 'zhiyu_qujs';
        $exchange = 'zhiyu_exjs';
        
        $post_data = [];
        $post_data['username'] = '888888';
        $post_data['money_type'] = 'money';
        $post_data['prize_type'] = 'jiandian';
        $post_data['number'] = '21.6';
        $post_data['money_edit'] = 'add';
        
        $msg_id = getMsgIds();
        
        $msg_data = [];
        $msg_data['MsgId'] = $msg_id;
        
        $msg_data['Service'] = 'WalletSync';
        $msg_data['Method'] = 'run';
        
        $msg_data['Args']['data'] = $post_data;
        
        //json序列化并且base64编码
        $msg_data = base64_encode(json_encode($msg_data));
        
        //构造保存数据
        $msg_save = ['msgid' => $msg_id, 'data' => $msg_data];
        
        Db::name('mq_msg_send')->insert($msg_save);
        
        $msg_publisher = new MsgConfirmPublisher();
        $msg_publisher->queue = $query;
        
        $result = '';
        while(!$result){
            $result = $msg_publisher->publisher($msg_data, $exchange, $query);
        }
        Db::name('mq_msg_send')->where('msgid', $msg_id)->update(['status'=>1]);
        return true;
    } */
}
?>