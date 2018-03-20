// Copyright (C) MongoDB, Inc. 2017-present.
//
// Licensed under the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License. You may obtain
// a copy of the License at http://www.apache.org/licenses/LICENSE-2.0
//
// Based on gopkg.in/mgo.v2/bson by Gustavo Niemeyer
// See THIRD-PARTY-NOTICES for original license terms.

package main

import (
	"context"
	"encoding/hex"
	"fmt"
	"math/rand"
	"sync/atomic"
	"time"

	"github.com/mongodb-labs/mgobson"
	"github.com/mongodb/mongo-go-driver/mongo"
)

var seed = time.Now().UnixNano()

func getCollectionName() string {
	atomic.AddInt64(&seed, 1)
	rand.Seed(atomic.LoadInt64(&seed))

	randomBytes := make([]byte, 16)
	_, err := rand.Read(randomBytes)
	if err != nil {
		panic(err)
	}

	return hex.EncodeToString(randomBytes)
}

func main() {
	client, err := mongo.NewClient("mongodb://localhost:27017")
	if err != nil {
		panic(err)
	}

	db := client.Database("examples")
	coll := db.Collection(getCollectionName())

	_, err = coll.InsertMany(
		context.Background(),
		[]interface{}{
			mgobson.M{"x": int32(1)},
			mgobson.D{{"x", int32(2)}},
			mgobson.RawD{{
				"x",
				mgobson.Raw{
					Kind: 0x10,
					Data: []byte{0x3, 0x0, 0x0, 0x0},
				},
			}},
		},
	)
	if err != nil {
		panic(err)
	}

	cursor, err := coll.Find(
		context.Background(),
		nil,
	)
	if err != nil {
		panic(err)
	}

	cursor.Next(context.Background())
	var r mgobson.RawD
	if err = cursor.Decode(&r); err != nil {
		panic(err)
	}
	fmt.Printf("%#v\n", r)

	cursor.Next(context.Background())
	var m mgobson.M
	if err = cursor.Decode(&m); err != nil {
		panic(err)
	}
	fmt.Printf("%#v\n", m)

	cursor.Next(context.Background())
	var d mgobson.D
	if err = cursor.Decode(&d); err != nil {
		panic(err)
	}
	fmt.Printf("%#v\n", d)
}
