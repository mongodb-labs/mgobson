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
	"github.com/mongodb/mongo-go-driver/bson"
	"github.com/mongodb/mongo-go-driver/mongo/connstring"
	"github.com/mongodb/mongo-go-driver/mongo/private/cluster"
	"github.com/mongodb/mongo-go-driver/mongo/private/ops"
	"github.com/mongodb/mongo-go-driver/mongo/readpref"
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
	cs, err := connstring.Parse("mongodb://localhost:27017")
	if err != nil {
		panic(err)
	}

	c, err := cluster.New(cluster.WithConnString(cs))
	if err != nil {
		panic(err)
	}

	s, err := c.SelectServer(
		context.Background(),
		readpref.Selector(readpref.Primary()),
		readpref.Primary(),
	)
	if err != nil {
		panic(err)
	}

	ns := ops.NewNamespace("examples", getCollectionName())

	_, err = ops.Insert(
		context.Background(),
		s,
		ns,
		[]*bson.Document{
			mgobson.M{"x": int32(1)}.MarshalBSONDocumentUnsafe(),
			mgobson.D{{"x", int32(2)}}.MarshalBSONDocumentUnsafe(),
			mgobson.RawD{{
				"x",
				mgobson.Raw{
					Kind: 0x10,
					Data: []byte{0x3, 0x0, 0x0, 0x0},
				},
			}}.MarshalBSONDocumentUnsafe(),
		},
	)
	if err != nil {
		panic(err)
	}

	pipeline := mgobson.DocsToArray([]interface{}{
		mgobson.M{
			"$group": mgobson.M{
				"_id": 1,
				"x": mgobson.M{
					"$push": "$x",
				},
			},
		},
		mgobson.D{
			{"$project", mgobson.D{
				{"_id", 0},
			}},
		},
	})

	cursor, err := ops.Aggregate(
		context.Background(),
		s,
		ns,
		pipeline,
		false,
	)
	if err != nil {
		panic(err)
	}

	cursor.Next(context.Background())
	var m mgobson.M
	if err = cursor.Decode(&m); err != nil {
		panic(err)
	}

	fmt.Println(m)
}
