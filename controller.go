package main

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/iamgak/go-task/models"
	"github.com/iamgak/go-task/pkg"
)

func (app *Application) ListTask(c *gin.Context) {
	filter := models.NewFilters(c)
	tasks, err := app.Model.TaskModelORM.TaskListing(c.Request.Context(), filter)
	if err != nil {
		app.Logger.Error(err.Error())
		if err == pkg.ErrNoRecord {
			app.ErrorJSONResponse(c.Writer, http.StatusNotFound, err.Error())
			return
		}
		app.ErrorJSONResponse(c.Writer, http.StatusInternalServerError, "Internal Server Error")
		return
	}

	c.JSON(http.StatusOK, tasks)
}

func (app *Application) TaskListingById(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		app.Logger.Error(err.Error())
		app.sendJSONResponse(c.Writer, http.StatusBadRequest, "Incorrect Input data provided")
		return
	}

	data, err := app.Model.TaskModelORM.TaskById(c.Request.Context(), id)
	if err != nil {
		app.Logger.Error(err.Error())
		if err == pkg.ErrNoRecord {
			app.ErrorJSONResponse(c.Writer, http.StatusNotFound, err.Error())
			return
		}
		app.ErrorJSONResponse(c.Writer, http.StatusInternalServerError, "Internal Server Error")
		return
	}

	c.JSON(http.StatusOK, data)
}

func (app *Application) UpdateTask(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		app.Logger.Error(err.Error())
		app.sendJSONResponse(c.Writer, http.StatusBadRequest, "Incorrect Input data provided")
		return
	}

	var task models.Task
	if err := c.ShouldBindJSON(&task); err != nil {
		app.Logger.Error("Loading Input Data Err :", err.Error())
		app.sendJSONResponse(c.Writer, http.StatusBadRequest, "Incorrect Input data provided")
		return
	}

	validator := app.Model.TaskModelORM.ValidateTaskData(&task, true)
	if len(validator.Errors) != 0 {
		c.JSON(http.StatusBadRequest, validator)
		return
	}

	task.UserID = app.UserID
	err = app.Model.TaskModelORM.UpdateTask(c.Request.Context(), id, &task)
	if err != nil {
		app.Logger.Error("error updating data ", err.Error())
		if err == pkg.ErrInvalidUserFound {
			app.ErrorJSONResponse(c.Writer, http.StatusNotFound, err.Error())
			return
		}
		app.ErrorJSONResponse(c.Writer, http.StatusInternalServerError, "Internal Server Error")
		return
	}

	activity := models.UserActivityLog{UserID: app.UserID, Activity: "Task Updated"}
	err = app.Model.UsersORM.UserActivityLog(&activity)
	if err != nil {
		app.Logger.Error(err.Error())
		app.sendJSONResponse(c.Writer, http.StatusBadRequest, "Internal Server Error")
		return
	}

	app.sendJSONResponse(c.Writer, http.StatusOK, task)
}

func (app *Application) UserActivateAccount(c *gin.Context) {
	token := c.Param("token")
	err := app.Model.UsersORM.ActivateAccount(token)
	if err != nil {
		app.Logger.Error(err.Error())
		if err == pkg.ErrNoRecord {
			app.ErrorJSONResponse(c.Writer, http.StatusNotFound, err.Error())
			return
		}

		app.ErrorJSONResponse(c.Writer, http.StatusInternalServerError, "Internal Server Error")
		return
	}

	app.sendJSONResponse(c.Writer, http.StatusOK, "Account Activated Successfully")
}

func (app *Application) SoftDelete(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		app.Logger.Error(err.Error())
		app.sendJSONResponse(c.Writer, http.StatusBadRequest, "Incorrect Input data provided")
		return
	}

	err = app.Model.TaskModelORM.SoftDelete(c.Request.Context(), app.UserID, uint(id))
	if err != nil {
		app.Logger.Error(err.Error())
		if err == pkg.ErrInvalidUserFound {
			app.ErrorJSONResponse(c.Writer, http.StatusBadRequest, err.Error())
			return
		}

		app.ErrorJSONResponse(c.Writer, http.StatusBadRequest, "Internal Server Error")
		return
	}

	activity := models.UserActivityLog{UserID: app.UserID, Activity: "Task Deleted"}
	err = app.Model.UsersORM.UserActivityLog(&activity)
	if err != nil {
		app.Logger.Error(err.Error())
		app.sendJSONResponse(c.Writer, http.StatusBadRequest, "Internal Server Error")
		return
	}
	app.sendJSONResponse(c.Writer, http.StatusOK, "Deleted Successfully")
}

func (app *Application) CreateTask(c *gin.Context) {
	var task models.Task
	if err := c.ShouldBindJSON(&task); err != nil {
		app.Logger.Error("Loading Input Data Err :", err.Error())
		app.sendJSONResponse(c.Writer, http.StatusBadRequest, "Incorrect Input data provided")
		return
	}

	validator := app.Model.TaskModelORM.ValidateTaskData(&task, false)
	if len(validator.Errors) != 0 {
		app.sendJSONResponse(c.Writer, http.StatusBadRequest, validator)
		return
	}

	task.UserID = app.UserID
	err := app.Model.TaskModelORM.CreateTask(c.Request.Context(), &task)
	if err != nil {
		app.Logger.Error(err.Error())
		app.sendJSONResponse(c.Writer, http.StatusBadRequest, "Internal Server Error")
		return
	}

	activity := models.UserActivityLog{UserID: task.UserID, Activity: "New Task Created"}
	err = app.Model.UsersORM.UserActivityLog(&activity)
	if err != nil {
		app.Logger.Error(err.Error())
		app.sendJSONResponse(c.Writer, http.StatusBadRequest, "Internal Server Error")
		return
	}
	app.sendJSONResponse(c.Writer, http.StatusCreated, task)
}

func (app *Application) UserLogin(c *gin.Context) {
	var creds *models.UserStruct
	if err := c.ShouldBindJSON(&creds); err != nil {
		app.Logger.Error("Loading Input Data Err :", err.Error())
		app.sendJSONResponse(c.Writer, http.StatusBadRequest, "Incorrect Input data provided")
		return
	}

	validator := app.Model.UsersORM.ValidateUserData(creds, false)
	if len(validator.Errors) != 0 {
		c.JSON(http.StatusBadRequest, validator)
		return
	}

	token, err := app.Model.UsersORM.LoginUser(c.Request.Context(), creds)
	if err != nil {
		app.Logger.Error(err.Error())
		if err == pkg.ErrAccountInActive {
			app.ErrorJSONResponse(c.Writer, http.StatusBadRequest, err.Error())
			return
		}

		if err == pkg.ErrInvalidCredentials {
			app.ErrorJSONResponse(c.Writer, http.StatusBadRequest, err.Error())
			return
		}

		app.ErrorJSONResponse(c.Writer, http.StatusInternalServerError, "Internal Server Error")
		return
	}

	c.Header("Authorization", "Bearer "+token)
	app.sendJSONResponse(c.Writer, http.StatusOK, "Login Successfull")
}

func (app *Application) UserRegister(c *gin.Context) {
	var creds *models.UserStruct
	if err := c.ShouldBindJSON(&creds); err != nil {
		app.Logger.Error("Loading Input Data Err :", err.Error())
		app.sendJSONResponse(c.Writer, http.StatusBadRequest, "Incorrect Input data provided")
		return
	}

	validator := app.Model.UsersORM.ValidateUserData(creds, true)
	if len(validator.Errors) != 0 {
		c.JSON(http.StatusBadRequest, validator)
		return
	}

	if err := app.Model.UsersORM.RegisterUser(c.Request.Context(), creds.Email, creds.Passw, c.ClientIP()); err != nil {
		app.Logger.Error(err.Error())
		app.sendJSONResponse(c.Writer, http.StatusBadRequest, "Internal Server Error")
		return
	}

	app.sendJSONResponse(c.Writer, http.StatusCreated, "Registration Successfully")
}
