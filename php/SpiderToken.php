<?php

namespace app\home\command;

use app\common\service\Base;
use function PHPSTORM_META\type;
use think\Db;
use Symfony\Component\DomCrawler\Crawler;
use think\console\Command;
use think\console\Input;
use think\console\input\Argument;
use think\console\Output;

/**
 * @author 基础小组
 * @group name=coin,description=货币资产
 */
class SpiderToken extends Command
{
    static $AUTH_CHECK = false;

    protected $rootUrl = 'https://etherscan.io';
    protected $rs =[];
    protected $beginTimeUnix;

    protected function configure()
    {
        $this->setName('SpiderToken')->setDescription('SpiderToken');
    }


    /**
     * @description 获取货币列表
     */
    protected function execute(Input $input, Output $output)
    {


        $beginTime = date('y-m-d H:i:s');
        echo "开始时间 $beginTime\n";
        $this->beginTimeUnix = time();

        for ($i = 10; $i <=10;$i++){

            $html = curl_get("https://etherscan.io/tokens?p={$i}",$this->getHeader());

            $crawler = new Crawler($html);

            $rs= $crawler
                ->filter('#ContentPlaceHolder1_divresult table tr')
                ->reduce(function (Crawler $node, $i) {
                    return $i>0;
                })
                ->each(function(Crawler $node, $i){
                    $temp= [];
                    $node->children()->each(function(Crawler $td,$j) use (&$temp){


                        switch ($j){
                            case 0:

                                $number = $td->filter('span')->text();
                                $temp['id'] = substr($number,2);
//                               if ($temp['id'] <=378){
//                                   return;
//                               }

                                break;
                            case 1: //logo
                                $path = $td->filter('img')->attr('src');
                                $imgPath = $this->download($this->rootUrl.$path);
                                $temp['img_path'] = $imgPath;
                                break;
                            case 2: //name
                                $a = $td->filter('h5>a');
                                $name = $a->text();

                                preg_match('/\((.*?)\)/',$name,$short_name);


                                $name = explode('(',$name);
                                $temp['full_name'] =$name[0];
                                $temp['short_name'] = $short_name[1];

                                $href = $this->rootUrl.$a->attr('href');
                                try{
                                    $rs = $this->getContract($href);
                                    $temp = array_merge($temp,$rs);
                                }catch (\Exception $e){
                                    return;
                                }

                                break;

                            case 3://price
                                break;
                            default:
                                break;
                        }


                    });

//                   if ($temp['id'] <=378){
//                       echo "{$temp['id']}退出\n";
//                       return;
//                   }


                    $spendTime = time()-$this->beginTimeUnix;
                    echo "第{$temp['id']}个完成,花费时间为{$spendTime}秒\n";

                    Db::name('token_info_copy')->insert($temp);
                    $temp=  [];

                });
        }

    }


    public function getContract($href){

        $html = curl_get($href,$this->getHeader());
        $crawler = new Crawler($html);
        $contract = $crawler->filter('table #ContentPlaceHolder1_trContract')->children()->eq(1)->children()->eq(0)->text();
        $decimal =  $crawler->filter('table #ContentPlaceHolder1_trContract')->siblings()->eq(1)->children()->eq(1)->text();

        //获取gas的连接
        $gas = $this->getGas($href);
        return compact('contract','gas','decimal');
    }

    public function getGas($href){
        $t = parse_url($href);
        $t = explode('/',$t['path']);
        $t = $t[count($t)-1];

        $gasUrl = $this->rootUrl.'/token/generic-tokentxns2?a=&mode=&contractAddress='.$t;
        $html = curl_get($gasUrl,$this->getHeader());

        $crawler = new Crawler($html);

        $gasHref = $crawler->filter('.table tr .address-tag')->eq(0)->children()->attr('href');
        $gasHref = $this->rootUrl.$gasHref;
        $html = curl_get($gasHref);
        $crawler = new Crawler($html);
        $gas = $crawler->filter('#ContentPlaceHolder1_maintable')->children()->eq(19)->children()->text();
        $gas = str_replace("\n",'',$gas);

        return $gas;

    }

    public function getHeader(){
        $header = <<<EOT
:authority:etherscan.io
:method:GET
:path:/tokens
:scheme:https
accept:text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8
accept-language:zh-CN,zh;q=0.9
cache-control:no-cache
pragma:no-cache
referer:https://etherscan.io/
upgrade-insecure-requests:1
user-agent:Mozilla/5.0 (Windows NT 10.0; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/63.0.3239.132 Safari/537.36
EOT;

        $header = explode("\r\n",$header);
        return $header;

    }

    function download($url)
    {
        $path = ROOT_PATH.'/public/coinImage/';
        $ch = curl_init();
        curl_setopt($ch, CURLOPT_URL, $url);
        curl_setopt($ch, CURLOPT_RETURNTRANSFER, 1);
        curl_setopt($ch, CURLOPT_CONNECTTIMEOUT, 30);
        $file = curl_exec($ch);
        curl_close($ch);
        $filename = pathinfo($url, PATHINFO_BASENAME);
        $resource = fopen($path . $filename, 'a');
        fwrite($resource, $file);
        fclose($resource);
        return '/public/coinImage/'.$filename;
    }


}

?>