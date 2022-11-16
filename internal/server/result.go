package server

import (
	"context"
	"fmt"
	"net/http"
	"quiz-app/internal/middleware"
	"quiz-app/internal/models"
	db "quiz-app/internal/mongoDB"
	"strings"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func AnswerQuestion(ctx *gin.Context) {
	fmt.Println(ctx.Request.Body, ": body")
	// questionID := ctx.Param("questionID")
	userAnswer := &models.UserAnswer{}

	if err := ctx.BindJSON(&userAnswer); err != nil {
		fmt.Println("error binding user answer: ", err)
		ctx.IndentedJSON(http.StatusExpectationFailed, err.Error())
		return
	}

	resultChannel := make(chan models.User)
	validUser := strings.TrimSpace(userAnswer.UserID)
	validQuestion := strings.TrimSpace(userAnswer.QuestionID)
	validOption := strings.TrimSpace(userAnswer.ChoosedOption)

	filter := bson.M{"userID": validUser}
	oldResult, ResultExists := middleware.ResultExists(validUser)

	go func() {

		foundUser, isUserFound := middleware.IsUserFound(validUser)
		if !isUserFound {
			ctx.IndentedJSON(http.StatusExpectationFailed, "user with this userID not found")
			return
		}

		resultChannel <- *foundUser
	}()
	if !ResultExists {
		ctx.IndentedJSON(http.StatusExpectationFailed, "result for this user not found")
		return

	}

	user := <-resultChannel
	oldResult.Points = oldResult.Points + userAnswer.Points
	data := &models.UserAnswer{
		UserID:        validUser,
		QuestionID:    validQuestion,
		ChoosedOption: validOption,
		Points:        oldResult.Points,
		Name:          user.Name,
		Email:         user.Email,
		Phone:         user.Phone,
	}

	_, err := db.ResultCollection.UpdateOne(db.MongoCtx, filter, bson.M{"$set": data})
	// fmt.Println("userAnswer: ", userAnswer)

	if err != nil {
		fmt.Println("error answering question: ", err)
		ctx.IndentedJSON(http.StatusInternalServerError, err.Error())
		return
	}

	ctx.IndentedJSON(http.StatusAccepted, data)

}

func ListUsersStanding(c *gin.Context) {
	userAnswer := []*models.UserAnswer{}

	filter := bson.D{}
	opts := options.Find().SetSort(bson.D{{"points", -1}})
	cursor, err := db.ResultCollection.Find(context.TODO(), filter, opts)
	if err != nil {

		fmt.Println("error listing users standing", err)
		c.IndentedJSON(http.StatusInternalServerError, err.Error())
		return
	}
	if err = cursor.All(context.TODO(), &userAnswer); err != nil {
		c.IndentedJSON(http.StatusInternalServerError, err.Error())
		return
	}

	c.IndentedJSON(http.StatusOK, userAnswer)
}
func DeleteResult(c *gin.Context) {
	userAnswer := []*models.UserAnswer{}
	userId := strings.TrimSpace(c.Param("userID"))

	_, resultExists := middleware.ResultExists(userId)

	if !resultExists {
		c.IndentedJSON(http.StatusExpectationFailed, "result with this userID not found")
		return
	}
	filter := bson.M{"userID": userId}

	err := db.DeleteOne(userAnswer, filter, db.ResultCollection)
	if err != nil {
		fmt.Println("error deleting user: ", err)
		c.IndentedJSON(http.StatusInternalServerError, err.Error())
		return
	}

	c.IndentedJSON(http.StatusAccepted, ("user deleted succesfully"))

}
