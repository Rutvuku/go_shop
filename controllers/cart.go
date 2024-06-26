package controllers

import (
	"context"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/Rutvuku/go_restro/database"
	"github.com/Rutvuku/go_restro/models"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type Application struct {
	prodCollection *mongo.Collection
	userCollection *mongo.Collection
}

func NewApplication(prodCollection *mongo.Collection, userCollection *mongo.Collection) *Application {
	return &Application{
		prodCollection: prodCollection,
		userCollection: userCollection,
	}
}

func (app *Application) AddToCart() gin.HandlerFunc {
	return func(c *gin.Context) {
		productQueryId := c.Query("id")
		if productQueryId == "" {
			log.Fatal("product id is empty")
			_ = c.AbortWithError(http.StatusBadRequest, errors.New("product id is empty"))
			return
		}
		userQueryId := c.Query("userID")
		if userQueryId == "" {
			log.Fatal("user id is empty")
			_ = c.AbortWithError(http.StatusBadRequest, errors.New("user id is empty"))
		}
		productID, err := primitive.ObjectIDFromHex(productQueryId)
		if err != nil {
			log.Fatal(err)
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		err = database.AddProductToCart(ctx, app.prodCollection, app.userCollection, productID, userQueryId)
		if err != nil {
			c.IndentedJSON(http.StatusInternalServerError, err)

		}
		c.IndentedJSON(200, "succesfully added to cart")
	}
}

func (app *Application) RemoveItem() gin.HandlerFunc {
	return func(c *gin.Context) {
		productQueryId := c.Query("id")
		if productQueryId == "" {
			log.Fatal("product id is empty")
			_ = c.AbortWithError(http.StatusBadRequest, errors.New("product id is empty"))
			return
		}
		userQueryId := c.Query("userID")
		if userQueryId == "" {
			log.Fatal("user id is empty")
			_ = c.AbortWithError(http.StatusBadRequest, errors.New("user id is empty"))
			return
		}
		productId, err := primitive.ObjectIDFromHex(productQueryId)
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		err = database.RemoveCartItem(ctx, app.prodCollection, app.userCollection, productId, userQueryId)
		if err != nil {
			c.IndentedJSON(http.StatusInternalServerError, err)
			return
		}
		c.IndentedJSON(200, "cart item succesfully removed")

	}
}

func GetItemFromCart() gin.HandlerFunc {
	return func(c *gin.Context) {
		user_id := c.Query("id")
		if user_id == "" {
			c.Header("Content-Type", "application/json")
			c.JSON(http.StatusNotFound, gin.H{"error": "invalid id"})
			c.Abort()
			return
		}
		usert_id, _ := primitive.ObjectIDFromHex(user_id)
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()
		var filledcart models.User
		err := UserCollection.FindOne(ctx, bson.D{primitive.E{Key: "_id", Value: usert_id}}).Decode(&filledcart)
		if err != nil {
			log.Println(err)
			c.IndentedJSON(500, "user not found")
			c.Abort()
			return
		}

		filter_match := bson.D{{Key: "$match", Value: bson.D{primitive.E{Key: "_id", Value: usert_id}}}}
		unwind := bson.D{{Key: "$unwind", Value: bson.D{primitive.E{Key: "path", Value: "$usercart"}}}}
		grouping := bson.D{{Key: "$group", Value: bson.D{primitive.E{Key: "_id", Value: "$_id"}, {Key: "total", Value: bson.D{primitive.E{Key: "$sum", Value: "$usercart.price"}}}}}}
		pointcursor, err := UserCollection.Aggregate(ctx, mongo.Pipeline{filter_match, unwind, grouping})
		if err != nil {
			c.IndentedJSON(http.StatusInternalServerError, "mongo error")
		}
		var listing []bson.M
		if err := pointcursor.All(ctx, &listing); err != nil {
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		for _, json := range listing {
			c.IndentedJSON(200, json["total"])
			c.IndentedJSON(200, filledcart.UserCart)
		}
		pointcursor.Close(ctx)
		ctx.Done()

	}
}

func (app *Application) BuyFromCart() gin.HandlerFunc {
	return func(c *gin.Context) {
		userQueryID := c.Query("id")
		if userQueryID == "" {
			log.Panicln("user id is empty")
			_ = c.AbortWithError(http.StatusBadRequest, errors.New("UserID is empty"))
		}
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()
		err := database.BuyItemFromCart(ctx, app.userCollection, userQueryID)
		if err != nil {
			c.IndentedJSON(http.StatusInternalServerError, err)
		}
		c.IndentedJSON(200, "Successfully Placed the order")
	}
}

func (app *Application) InstantBuy() gin.HandlerFunc {
	return func(c *gin.Context) {
		productQueryId := c.Query("id")
		if productQueryId == "" {
			log.Fatal("product id is empty")
			_ = c.AbortWithError(http.StatusBadRequest, errors.New("product id is empty"))
			return
		}
		userQueryId := c.Query("userID")
		if userQueryId == "" {
			log.Fatal("user id is empty")
			_ = c.AbortWithError(http.StatusBadRequest, errors.New("user id is empty"))
			return
		}
		productId, err := primitive.ObjectIDFromHex(productQueryId)
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		err = database.InstantBuyer(ctx, app.prodCollection, app.userCollection, productId, userQueryId)
		if err != nil {
			c.IndentedJSON(http.StatusInternalServerError, err)
			return
		}
		c.IndentedJSON(200, "cart item bought instantly")

	}
}
