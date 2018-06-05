package main

import (
	"bytes"
	"log"

	"fmt"

	"github.com/boltdb/bolt"
)

func base(db *bolt.DB) {
	db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte("a_table")) //创建并获取bucket，bucket类似于表

		if err != nil {
			return fmt.Errorf("create bucket:%s", err)
		}
		b := tx.Bucket([]byte("a_table")) //获取bucket

		err = b.Put([]byte("name"), []byte("mtg")) // 存储键值对到a_table
		err = b.Put([]byte("na"), []byte("m"))     // 存储键值对到a_table
		err = b.Put([]byte("user"), []byte("wjx")) // 存储键值对到a_table

		v := b.Get([]byte("name")) //如果键存在，它将返回它的字节切片值。如果它不存在，它将返回nil
		fmt.Println(string(v))

		return nil
	})

}

//普通遍历
func cursor(db *bolt.DB) {
	db.View(func(tx *bolt.Tx) error {
		// Assume bucket exists and has keys
		b := tx.Bucket([]byte("a_table"))

		c := b.Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {
			fmt.Printf("key=%s, value=%s\n", k, v)
		}

		return nil
	})
}

//前缀扫描
func prefix(db *bolt.DB) {
	db.View(func(tx *bolt.Tx) error {
		// Assume bucket exists and has keys
		c := tx.Bucket([]byte("a_table")).Cursor()

		prefix := []byte("n")

		for k, v := c.Seek(prefix); k != nil && bytes.HasPrefix(k, prefix); k, v = c.Next() {
			fmt.Printf("key=%s, value=%s\n", k, v)
		}

		return nil
	})
}

//foreach 遍历
func each(db *bolt.DB) {
	db.View(func(tx *bolt.Tx) error {
		// Assume bucket exists and has keys
		b := tx.Bucket([]byte("a_table"))

		b.ForEach(func(k, v []byte) error {
			fmt.Printf("key=%s, value=%s\n", k, v)
			return nil
		})
		return nil
	})
}

func main() {
	// Open the my.db data file in your current directory.
	// It will be created if it doesn't exist.
	db, err := bolt.Open("my.db", 0600, nil)
	base(db)
	cursor(db)
	fmt.Println("=======")
	prefix(db)
	fmt.Println("=======")
	each(db)
	if err != nil {
		log.Fatal(err)
	}

	defer db.Close()

}
