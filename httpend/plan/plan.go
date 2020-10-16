// Package plan is responsible for all endpoints on the /plans resource
package plan

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/gin-gonic/gin"
	"github.com/wallnutkraken/groupplan/groupdata/dataerror"
	"github.com/wallnutkraken/groupplan/httpend/shtypes"
	"github.com/wallnutkraken/groupplan/httpend/userauth"
	"github.com/wallnutkraken/groupplan/planman"
)

// Handler is the object responsible for the /plans endpoint
type Handler struct {
	group   *gin.RouterGroup
	auther  userauth.Authenticator
	planner planman.Planner
}

// New creates a new instance of the plans handler
func New(router *gin.Engine, auth userauth.Authenticator, planner planman.Planner) *Handler {
	handl := &Handler{
		group:   router.Group("plans"),
		auther:  auth,
		planner: planner,
	}

	// Add the endpoints
	handl.group.PUT("", handl.NewPlan)
	handl.group.GET(":identifier", handl.GetPlan)
	handl.group.GET("", handl.MyPlans)
	handl.group.PUT(":identifier", handl.AddEntry)
	handl.group.DELETE(":identifier", handl.DeletePlan)
	handl.group.DELETE(":identifier/entries/:entryID", handl.DeleteEntry)
	handl.group.GET(":identifier/entries", handl.GetEntriesForPlan)

	return handl
}

// NewPlan creates a new plan
func (h Handler) NewPlan(ctx *gin.Context) {
	// Check authorization
	user, err := h.auther.GetJWT(ctx)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, shtypes.NewUserError("Please log in"))
		return
	}

	// Parse the request body, add a default value for MinAvailabilitySeconds
	// at 5 minutes.
	req := CreatePlanRequest{
		MinAvailabilitySeconds: 60 * 5,
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.AbortWithStatusJSON(http.StatusUnprocessableEntity, shtypes.NewUserError(err.Error()))
		return
	}
	// Parse the start date
	startDate, err := time.Parse("2006-1-2", req.StartDate)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusUnprocessableEntity, shtypes.NewUserError(fmt.Sprintf("invalid start date format [%s], please use yyyy-mm-dd", req.StartDate)))
		return
	}
	plan, err := h.planner.NewPlan(req.Title, startDate, req.DurationDays, user)
	if err != nil {
		if errors.As(err, &dataerror.BaseError{}) {
			// User error, return the contents with an error
			ctx.AbortWithStatusJSON(http.StatusUnprocessableEntity, shtypes.NewUserError(err.Error()))
			return
		}
		// Non-user error, log it and return 500
		refErr := shtypes.NewServerError()
		logrus.WithError(err).Errorf("[%s]", refErr.Reference)
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, refErr)
		return
	}

	// Plan created, return it
	ctx.JSON(http.StatusCreated, plan)
}

// GetPlan gets a plan with a given identifier
func (h Handler) GetPlan(ctx *gin.Context) {
	identifier := ctx.Param("identifier")

	// Check authorization
	_, err := h.auther.GetJWT(ctx)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, shtypes.NewUserError("Please log in"))
		return
	}

	plan, err := h.planner.GetPlan(identifier)
	if err != nil {
		if errors.As(err, &dataerror.BaseError{}) {
			// User error, return the contents with an error
			ctx.AbortWithStatusJSON(http.StatusUnprocessableEntity, shtypes.NewUserError(err.Error()))
			return
		}
		// Non-user error, log it and return 500
		refErr := shtypes.NewServerError()
		logrus.WithError(err).Errorf("[%s]", refErr.Reference)
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, refErr)
		return
	}

	// We got the plan, return it
	ctx.JSON(http.StatusOK, plan)
}

// MyPlans returns the authorized user's owned plans
func (h Handler) MyPlans(ctx *gin.Context) {
	// Check authorization
	user, err := h.auther.GetJWT(ctx)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, shtypes.NewUserError("Please log in"))
		return
	}
	// Get the plans for the authorized user
	plans, err := h.planner.GetPlans(user)
	if err != nil {
		// There can be no user input error, so we just log whatever this error is
		// and return an internal server error
		refErr := shtypes.NewServerError()
		logrus.WithError(err).Errorf("[%s]", refErr.Reference)
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, refErr)
		return
	}

	// Got the plan list, return it to the consumer
	ctx.JSON(http.StatusOK, plans)
}

// AddEntry adds a new availability entry to a given plan
func (h Handler) AddEntry(ctx *gin.Context) {
	// Check authorization
	user, err := h.auther.GetJWT(ctx)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, shtypes.NewUserError("Please log in"))
		return
	}
	// Read the request body
	req := AddEntryRequest{}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.AbortWithStatusJSON(http.StatusUnprocessableEntity, shtypes.NewUserError(err.Error()))
		return
	}

	entry, err := h.planner.AddEntry(ctx.Param("identifier"), user, req.StartTime, req.DurationSeconds)
	if err != nil {
		if errors.As(err, &dataerror.BaseError{}) {
			ctx.AbortWithStatusJSON(http.StatusUnprocessableEntity, shtypes.NewUserError(err.Error()))
			return
		}
		// Non-user error, log it and return 500
		refErr := shtypes.NewServerError()
		logrus.WithError(err).Errorf("[%s]", refErr.Reference)
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, refErr)
		return
	}

	// Entry created, return the created object to the user
	ctx.JSON(http.StatusCreated, entry)
}

// DeletePlan deletes an existing plan
func (h Handler) DeletePlan(ctx *gin.Context) {
	// Check authorization
	user, err := h.auther.GetJWT(ctx)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, shtypes.NewUserError("Please log in"))
		return
	}
	// Get the identifier
	identifier := ctx.Param("identifier")

	if err := h.planner.DeletePlan(identifier, user); err != nil {
		userError := dataerror.BaseError{}
		if errors.As(err, &userError) {
			// User error, return the contents with an error
			ctx.AbortWithStatusJSON(userError.StatusCode(), shtypes.NewUserError(userError.Error()))
			return
		}
		// Non-user error, log it and return 500
		refErr := shtypes.NewServerError()
		logrus.WithError(err).Errorf("[%s]", refErr.Reference)
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, refErr)
		return
	}

	ctx.Status(http.StatusNoContent)
}

// DeleteEntry is the endpoint handling the deletion of an availability entry for a plan
func (h Handler) DeleteEntry(ctx *gin.Context) {
	// Check authorization
	user, err := h.auther.GetJWT(ctx)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, shtypes.NewUserError("Please log in"))
		return
	}
	// Here's the deal, this endpoint is in /plans/:planID/entries/:entryID
	// Why does it have a :planID even if it doesn't care about plans, just the entries?
	// Because of gin's router. If I set up a route like DELETE /plans/entries/:entryID
	// It will cause a panic on launch, just the way the router works. So don't blame me,
	// please.

	// Get the entryID, parse into uint
	entryIDString := ctx.Param("entryID")
	entryID, err := strconv.ParseUint(entryIDString, 10, 32)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, shtypes.NewUserError("entry ID is not an unsigned integer"))
		return
	}

	if err := h.planner.DeleteEntry(uint(entryID), user); err != nil {
		userError := dataerror.BaseError{}
		if errors.As(err, &userError) {
			// User error, return the contents with an error
			ctx.AbortWithStatusJSON(userError.StatusCode(), shtypes.NewUserError(userError.Error()))
			return
		}
		// Non-user error, log it and return 500
		refErr := shtypes.NewServerError()
		logrus.WithError(err).Errorf("[%s]", refErr.Reference)
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, refErr)
		return
	}

	// No errors, return a No Content
	ctx.Status(http.StatusNoContent)
}

// GetEntriesForPlan returns the user's only the entries for the plan requested.
func (h Handler) GetEntriesForPlan(ctx *gin.Context) {
	// Check authorization
	user, err := h.auther.GetJWT(ctx)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, shtypes.NewUserError("Please log in"))
		return
	}

	identifier := ctx.Param("identifier")
	entries, err := h.planner.GetEntriesOnPlanByUser(identifier, user)
	if err != nil {
		// There can be no user input error, so we just log whatever this error is
		// and return an internal server error
		refErr := shtypes.NewServerError()
		logrus.WithError(err).Errorf("[%s]", refErr.Reference)
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, refErr)
		return
	}

	ctx.JSON(http.StatusOK, entries)
}
