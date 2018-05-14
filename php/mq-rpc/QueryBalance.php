<?php
use think\Db;

/**
 *  查询某个币的余额
 */
class QueryBalance {


    public static function run($data){
        $symbol_id = $data['data']['symbol_id'];

        $coinPrice = logic('Trade')->getNowPrice($symbol_id);

        $coinName = Db::name('coin_symbol')->where('id',$symbol_id)->value('title');

        return [
            'status'=>true,
             'data'=> compact('coinName','coinPrice')
        ];
    }
}