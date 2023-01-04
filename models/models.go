package models

import (
	"github.com/go-redis/redis/v8"
)

var (
	Rdb *redis.Client
)

/*	redis指令

	// err := rdb.Set(ctx, "key", "666666", 0).Err()
    // if err != nil {
    //     panic(err)
    // }

    // val, err := rdb.Get(ctx, "key").Result()
    // if err != nil {
    //     panic(err)
    // }
    // fmt.Println("key", val)

    // val2, err := rdb.Get(ctx, "key2").Result()
    // if err == redis.Nil {
    //     fmt.Println("key2 does not exist")
    // } else if err != nil {
    //     panic(err)
    // } else {
    //     fmt.Println("key2", val2)
    // }
    // Output: key value
    // key2 does not exist

	//set hash 適合儲存結構
	// rdb.HSet(ctx, "user", "key1", "value1", "key2", "value2")
	// rdb.Del(ctx, "user")
	// rdb.HMSet(ctx, "user", map[string]interface{}{"name":"kevin", "age": 27, "address":"北京"})

	// //HGet():获取某个元素
	// address, err := rdb.HGet(ctx, "user", "address").Result()
	// if err != nil {
	// 	fmt.Println("addreersss")
	// 	fmt.Println(err)
	// 	panic(err)
	// }
	// fmt.Println(address)

	// //HGetAll():获取全部元素
	// user, err := rdb.HGetAll(ctx, "user").Result()
	// if err != nil {
	// 	panic(err)
	// }
	// fmt.Println(user)

	// //HDel():删除某个元素
	// res, err := rdb.HDel(ctx, "user", "name", "age").Result()
	// if err != nil {
	// 	panic(err)
	// }
	// fmt.Println(res)

*/
