package ahandlers

import (
	"github.com/dropbox/godropbox/container/set"
	"github.com/gin-gonic/gin"
	"github.com/pritunl/pritunl-cloud/database"
	"github.com/pritunl/pritunl-cloud/demo"
	"github.com/pritunl/pritunl-cloud/event"
	"github.com/pritunl/pritunl-cloud/utils"
	"github.com/pritunl/pritunl-cloud/zone"
	"gopkg.in/mgo.v2/bson"
)

type zoneData struct {
	Id            bson.ObjectId   `json:"id"`
	Datacenter    bson.ObjectId   `json:"datacenter"`
	Organizations []bson.ObjectId `json:"organizations"`
	Name          string          `json:"name"`
}

func zonePut(c *gin.Context) {
	if demo.Blocked(c) {
		return
	}

	db := c.MustGet("db").(*database.Database)
	data := &zoneData{}

	zoneId, ok := utils.ParseObjectId(c.Param("zone_id"))
	if !ok {
		utils.AbortWithStatus(c, 400)
		return
	}

	err := c.Bind(data)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	zne, err := zone.Get(db, zoneId)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	zne.Name = data.Name
	zne.Organizations = data.Organizations

	fields := set.NewSet(
		"name",
		"organizations",
	)

	errData, err := zne.Validate(db)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	if errData != nil {
		c.JSON(400, errData)
		return
	}

	err = zne.CommitFields(db, fields)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	event.PublishDispatch(db, "zone.change")

	c.JSON(200, zne)
}

func zonePost(c *gin.Context) {
	if demo.Blocked(c) {
		return
	}

	db := c.MustGet("db").(*database.Database)
	data := &zoneData{
		Name: "New Zone",
	}

	err := c.Bind(data)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	zne := &zone.Zone{
		Datacenter:    data.Datacenter,
		Organizations: data.Organizations,
		Name:          data.Name,
	}

	errData, err := zne.Validate(db)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	if errData != nil {
		c.JSON(400, errData)
		return
	}

	err = zne.Insert(db)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	event.PublishDispatch(db, "zone.change")

	c.JSON(200, zne)
}

func zoneDelete(c *gin.Context) {
	if demo.Blocked(c) {
		return
	}

	db := c.MustGet("db").(*database.Database)

	zoneId, ok := utils.ParseObjectId(c.Param("zone_id"))
	if !ok {
		utils.AbortWithStatus(c, 400)
		return
	}

	err := zone.Remove(db, zoneId)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	event.PublishDispatch(db, "zone.change")

	c.JSON(200, nil)
}

func zoneGet(c *gin.Context) {
	db := c.MustGet("db").(*database.Database)

	zoneId, ok := utils.ParseObjectId(c.Param("zone_id"))
	if !ok {
		utils.AbortWithStatus(c, 400)
		return
	}

	zne, err := zone.Get(db, zoneId)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	c.JSON(200, zne)
}

func zonesGet(c *gin.Context) {
	db := c.MustGet("db").(*database.Database)

	zones, err := zone.GetAll(db)
	if err != nil {
		utils.AbortWithError(c, 500, err)
		return
	}

	c.JSON(200, zones)
}
