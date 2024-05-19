package controllers

import (
	"context"
	"net/http"
	"time"

	"github.com/Rutvuku/go_restro/models"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func AddAddress() gin.HandlerFunc {
	return func(c *gin.Context) {
		user_id := c.Query("id")
		if user_id == "" {
			c.Header("Content-Type,application/json")
			c.IndentedJSON(http.StatusBadRequest, gin.H{"Error": "invalid search id"})
			c.Abort()
			return
		}
		address, err := primitive.ObjectIDFromHex(user_id)
		if err != nil {
			c.IndentedJSON(500, "internal server error")
			return
		}
		var addresses models.Address
		addresses.Address_id = primitive.NewObjectID()
		if err := c.BindJSON(&addresses); err != nil {
			c.IndentedJSON(http.StatusInternalServerError, "internal server error")
			return
		}
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		match_filter := bson.D{{Key: "$match", Value: bson.D{primitive.E{Key: "_id", Value: address}}}}
		unwind := bson.D{{Key: "$unwind", Value: bson.D{primitive.E{Key: "path", Value: "$address"}}}}
		grouping := bson.D{{Key: "group", Value: bson.D{primitive.E{Key: "_id", Value: "$address_id"}, {Key: "count", Value: bson.D{primitive.E{Key: "$sum", Value: 1}}}}}}
		pointcursor, err := UserCollection.Aggregate(ctx, mongo.Pipeline{match_filter, unwind, grouping})
		if err != nil {
			c.IndentedJSON(500, "internal server error")
			return
		}

	}
}

func EditHomeAddress() gin.HandlerFunc {

}

func EditWorkAddress() gin.HandlerFunc {

}

func DeleteAddress() gin.HandlerFunc {
	return func(c *gin.Context) {
		var user_id = c.Query("id")
		if user_id == "" {
			c.Header("Content-Type,application/json")
			c.IndentedJSON(http.StatusBadRequest, gin.H{"Error": "invalid search id"})
			c.Abort()
			return
		}
		addresses := make([]models.Address, 0)
		usert_id, err := primitive.ObjectIDFromHex(user_id)
		if err != nil {
			c.IndentedJSON(500, "internal server error")
			return
		}
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()
		filter := bson.D{primitive.E{Key: "_id", Value: usert_id}}
		update := bson.D{primitive.E{Key: "$set", Value: bson.D{primitive.E{Key: "address", Value: addresses}}}}
		_, err = ProductCollection.UpdateOne(ctx, filter, update)
		if err != nil {
			c.IndentedJSON(404, "server error")
			return
		}
		defer cancel()
		ctx.Done()
		c.IndentedJSON(200, "successfully deleted")

	}
}
