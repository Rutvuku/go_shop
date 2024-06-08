package database

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/Rutvuku/go_restro/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

var (
	ErrCantFindProduct    = errors.New("can't find product")
	ErrCantDecodeProducts = errors.New("can't find product")
	ErrUserIDIsNotValid   = errors.New("user is not valid")
	ErrCantUpdateUser     = errors.New("cannot add product to cart")
	ErrCantRemoveItem     = errors.New("cannot remove item from cart")
	ErrCantGetItem        = errors.New("cannot get item from cart ")
	ErrCantBuyCartItem    = errors.New("cannot update the purchase")
)

func AddProductToCart(ctx context.Context, prodCollection *mongo.Collection, userCollection *mongo.Collection, productID primitive.ObjectID, userID string) error {
	searchfromdb, err := prodCollection.Find(ctx, bson.M{"_id": productID})
	if err != nil {
		log.Println(err)
		return ErrCantFindProduct
	}
	var productcart []models.ProductUser
	err = searchfromdb.All(ctx, &productcart)
	if err != nil {
		log.Fatal(err)
		return ErrCantDecodeProducts
	}
	id, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		log.Fatal(err)
		return ErrUserIDIsNotValid
	}
	filter := bson.D{primitive.E{Key: "_id", Value: id}}
	update := bson.D{{Key: "$push", Value: bson.D{primitive.E{Key: "usercart", Value: bson.D{primitive.E{Key: "$each", Value: productcart}}}}}}
	_, err = userCollection.UpdateOne(ctx, filter, update)
	if err != nil {
		return ErrCantUpdateUser
	}
	return nil
}

func RemoveCartItem(ctx context.Context, prodCollection *mongo.Collection, userCollection *mongo.Collection, productID primitive.ObjectID, userID string) error {
	id, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		log.Fatal(err)
		return ErrUserIDIsNotValid
	}
	filter := bson.D{primitive.E{Key: "_id", Value: id}}
	update := bson.M{"$pull": bson.M{"usercart": bson.M{"_id": productID}}}
	_, err = userCollection.UpdateMany(ctx, filter, update)
	if err != nil {
		log.Fatal(err)
		return ErrCantRemoveItem
	}
	return nil

}

func BuyItemFromCart(ctx context.Context, userCollection *mongo.Collection, userID string) error {
	id, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		//panic(err)
		return ErrUserIDIsNotValid
	}
	var getCartItems models.User
	var orderCart models.Order

	orderCart.Order_ID = primitive.NewObjectID()
	orderCart.Order_Cart = make([]models.ProductUser, 0)
	orderCart.Orderered_At = time.Now()
	orderCart.Payment_Method.COD = true

	unwind := bson.D{{Key: "$unwind", Value: bson.D{primitive.E{Key: "path", Value: "$usercart"}}}}
	grouping := bson.D{{Key: "$group", Value: bson.D{primitive.E{Key: "_id", Value: "$_id"}, {Key: "total", Value: bson.D{primitive.E{Key: "$sum", Value: "$usercart.price"}}}}}}
	currentresults, err := userCollection.Aggregate(ctx, mongo.Pipeline{unwind, grouping})
	if err != nil {
		panic(err)
	}
	ctx.Done()
	var getusercart []bson.M
	if err = currentresults.All(ctx, &getusercart); err != nil {
		panic(err)
	}
	var total_price int32
	for _, user_item := range getusercart {
		price := user_item["total"]
		total_price = price.(int32)
	}
	orderCart.Price = int(total_price)
	filter := bson.D{{Key: "_id", Value: "$_id"}}
	update := bson.D{{Key: "$push", Value: primitive.E{Key: "orders", Value: orderCart}}}
	if _, err := userCollection.UpdateMany(ctx, filter, update); err != nil {
		panic(err)
	}

	err = userCollection.FindOne(ctx, bson.D{{Key: "_id", Value: id}}).Decode(&getCartItems)
	if err != nil {
		panic(err)
	}
	filter2 := bson.D{{Key: "_id", Value: id}}
	update2 := bson.M{"$push": bson.M{"orders.$[].order_list": bson.M{"$each": getCartItems.UserCart}}}
	_, err = userCollection.UpdateOne(ctx, filter2, update2)
	if err != nil {
		panic(err)
	}
	usercart_empty := make([]models.ProductUser, 0)
	filter3 := bson.D{{Key: "_id", Value: id}}
	update3 := bson.D{{Key: "$set", Value: primitive.E{Key: "usercart", Value: usercart_empty}}}
	_, err = userCollection.UpdateOne(ctx, filter3, update3)
	if err != nil {
		panic(err)
	}
	return nil
}

// func InstantBuyer(ctx context.Context, prodCollection, userCollection *mongo.Collection, productID primitive.ObjectID, UserID string) error {
// 	id, err := primitive.ObjectIDFromHex(UserID)
// 	if err != nil {
// 		log.Println(err)
// 		return ErrUserIDIsNotValid
// 	}
// 	var getcartitems models.User
// 	var ordercart models.Order
// 	ordercart.Order_ID = primitive.NewObjectID()
// 	ordercart.Orderered_At = time.Now()
// 	ordercart.Order_Cart = make([]models.ProductUser, 0)
// 	ordercart.Payment_Method.COD = true
// 	unwind := bson.D{{Key: "$unwind", Value: bson.D{primitive.E{Key: "path", Value: "$usercart"}}}}
// 	grouping := bson.D{{Key: "$group", Value: bson.D{primitive.E{Key: "_id", Value: "$_id"}, {Key: "total", Value: bson.D{primitive.E{Key: "$sum", Value: "$usercart.price"}}}}}}
// 	currentresults, err := userCollection.Aggregate(ctx, mongo.Pipeline{unwind, grouping})
// 	ctx.Done()
// 	if err != nil {
// 		panic(err)
// 	}
// 	var getusercart []bson.M
// 	if err = currentresults.All(ctx, &getusercart); err != nil {
// 		panic(err)
// 	}
// 	var total_price int32
// 	for _, user_item := range getusercart {
// 		price := user_item["total"]
// 		total_price = price.(int32)
// 	}
// 	ordercart.Price = int(total_price)
// 	filter := bson.D{primitive.E{Key: "_id", Value: id}}
// 	update := bson.D{{Key: "$push", Value: bson.D{primitive.E{Key: "orders", Value: ordercart}}}}
// 	_, err = userCollection.UpdateMany(ctx, filter, update)
// 	if err != nil {
// 		log.Println(err)
// 	}
// 	err = userCollection.FindOne(ctx, bson.D{primitive.E{Key: "_id", Value: id}}).Decode(&getcartitems)
// 	if err != nil {
// 		log.Println(err)
// 	}
// 	filter2 := bson.D{primitive.E{Key: "_id", Value: id}}
// 	update2 := bson.M{"$push": bson.M{"orders.$[].order_list": bson.M{"$each": getcartitems.UserCart}}}
// 	_, err = userCollection.UpdateOne(ctx, filter2, update2)
// 	if err != nil {
// 		log.Println(err)
// 	}
// 	usercart_empty := make([]models.ProductUser, 0)
// 	filtered := bson.D{primitive.E{Key: "_id", Value: id}}
// 	updated := bson.D{{Key: "$set", Value: bson.D{primitive.E{Key: "usercart", Value: usercart_empty}}}}
// 	_, err = userCollection.UpdateOne(ctx, filtered, updated)
// 	if err != nil {
// 		return ErrCantBuyCartItem

// 	}
// 	return nil
// }

func InstantBuyer(ctx context.Context, prodCollection, userCollection *mongo.Collection, productID primitive.ObjectID, UserID string) error {
	id, err := primitive.ObjectIDFromHex(UserID)
	if err != nil {
		log.Println(err)
		return ErrUserIDIsNotValid
	}
	var product_details models.ProductUser
	var orders_detail models.Order
	orders_detail.Order_ID = primitive.NewObjectID()
	orders_detail.Orderered_At = time.Now()
	orders_detail.Order_Cart = make([]models.ProductUser, 0)
	orders_detail.Payment_Method.COD = true
	err = prodCollection.FindOne(ctx, bson.D{primitive.E{Key: "_id", Value: productID}}).Decode(&product_details)
	if err != nil {
		log.Println(err)
	}
	orders_detail.Price = product_details.Price
	filter := bson.D{primitive.E{Key: "_id", Value: id}}
	update := bson.D{{Key: "$push", Value: bson.D{primitive.E{Key: "orders", Value: orders_detail}}}}
	_, err = userCollection.UpdateOne(ctx, filter, update)
	if err != nil {
		log.Println(err)
	}
	filter2 := bson.D{primitive.E{Key: "_id", Value: id}}
	update2 := bson.M{"$push": bson.M{"orders.$[].order_list": product_details}}
	_, err = userCollection.UpdateOne(ctx, filter2, update2)
	if err != nil {
		log.Println(err)
	}
	return nil
}
