package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

type Album struct {
	ID     primitive.ObjectID `bson:"_id,omitempty"`
	Artist string             `bson:"artist,omitempty"`
	Title  string             `bson:"album,omitempty"`
	Year   int                `bson:"year,omitempty"`
}

type Albums []Album

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	opt := options.Client().ApplyURI("mongodb://localhost:27017")
	client, err := mongo.Connect(ctx, opt)
	if err != nil {
		log.Fatal(err)
	}
	if err := client.Ping(ctx, readpref.Primary()); err != nil {
		log.Fatal(err)
	}
	myAlbums := client.Database("mydb").Collection("albums")

	// Create
	entry := bson.D{{"artist", "Rammstein"}, {"album", "Zeit"}, {"year", 2022}}
	result, err := myAlbums.InsertOne(ctx, entry)
	if err != nil {
		log.Printf("could not insert entry %v: %v", entry, err)
	}
	fmt.Printf("mongo-ID: %v for %v\n", result.InsertedID, entry)

	entries := []interface{}{
		bson.D{{"artist", "Queen"}, {"album", "A Day at the Races"}, {"year", 1976}},
		bson.D{{"artist", "Beethoven"}, {"album", "9. Symphonie"}, {"year", 1824}},
	}
	results, err := myAlbums.InsertMany(ctx, entries)
	if err != nil {
		log.Printf("could not insert entries %v: %v", entries, err)
	}
	for i, id := range results.InsertedIDs {
		fmt.Printf("mongo-ID: %v for %v\n", id, entries[i])
	}

	// Read
	filter := bson.D{
		{"$and", bson.A{
			bson.D{{"year", bson.D{{"$gt", 1950}}}}}},
	}
	cursor, err := myAlbums.Find(ctx, filter)
	if err != nil {
		log.Printf("could not read form db: %v", err)
	}
	var res []bson.M // Alle Elemente in Filter
	if err = cursor.All(ctx, &res); err != nil {
		log.Printf("could convert from bson: %v", err)
	}
	for _, album := range res {
		fmt.Println(album)
	}
	var resFirst bson.M // erstes Element für Filter
	if err = myAlbums.FindOne(ctx, filter).Decode(&resFirst); err != nil {
		log.Printf("did not find entry: %v", err)
	}
	fmt.Println(resFirst)

	// leerer Filter = alle Elemente
	cursor, err = myAlbums.Find(ctx, bson.D{})
	if err != nil {
		log.Printf("no DB-entry found: %v", err)
	}
	fmt.Println()
	var albums Albums
	if err = cursor.All(ctx, &albums); err != nil {
		log.Printf("could not convert entries: %v", err)
	}
	for _, album := range albums {
		fmt.Println(album)
	}

	rammstein := Album{
		Artist: "ramstein",
		Title:  "Rammstein",
		Year:   2019,
	}
	rs, err := myAlbums.InsertOne(ctx, rammstein)
	if err != nil {
		log.Printf("not inserted %v: %v", rammstein, err)
	}
	update := bson.D{
		{"$set", bson.D{{"artist", "Rammstein"}}},
		{"$inc", bson.D{{"year", 1}}},
	}
	urs, err := myAlbums.UpdateByID(ctx, rs.InsertedID, update)
	if err != nil {
		log.Printf("could not update %v: %v", rs, err)
	}
	fmt.Printf("%d documents were updated\n", urs.ModifiedCount)

	// Delete
	delFilter := bson.D{
		{"$and", bson.A{bson.D{{"year", bson.D{{"$lt", 1950}}}}}},
	}
	resDel, err := myAlbums.DeleteMany(ctx, delFilter)
	if err != nil {
		log.Printf("could not delete %v: %v", filter, err)
	}
	fmt.Printf("%d documents were deleted\n", resDel.DeletedCount)

	cursor, err = myAlbums.Find(ctx, bson.D{})
	if err != nil {
		log.Printf("no DB-entry found: %v", err)
	}
	if err = cursor.All(ctx, &albums); err != nil {
		log.Printf("could not convert entries: %v", err)
	}
	fmt.Println()
	for _, album := range albums {
		fmt.Println(album)
	}
	myAlbums.Drop(ctx)
}
