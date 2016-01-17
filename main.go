package galwaybus

import (
        //"fmt"
        "net/http"
        "os"
        "strings"
        "encoding/json"
        "strconv"

	      "sqbu-github.cisco.com/jgoecke/go-spark"

        "github.com/go-martini/martini"
	      "github.com/martini-contrib/render"
	      "github.com/martini-contrib/binding"

        "google.golang.org/appengine"
      	"google.golang.org/appengine/log"
        "google.golang.org/appengine/urlfetch"

        "github.com/parnurzeal/gorequest"

)


type SparkEvent struct {
	Id          string `json:"id" binding:"required"`
	RoomId      string `json:"roomId" binding:"required"`
	PersonId    string `json:"personId" binding:"required"`
	PersonEmail string `json:"personEmail" binding:"required"`
	Text        string `json:"text" binding:"required"`
}

type BusRoute struct {
  Id          int `json:"timetable_id" binding:"required"`
	LongName    string `json:"long_name" binding:"required"`
  ShortName   string `json:"short_name" binding:"required"`

}


// init is called before the application starts.
func init() {

  m := martini.Classic()
  m.Use(render.Renderer())

  m.Use(func(res http.ResponseWriter, req *http.Request) {
		authorization := &spark.Authorization{AccessToken: os.Getenv("SPARK_TOKEN")}
		spark.InitClient(authorization)

		ctx := appengine.NewContext(req)
		spark.SetHttpClient(urlfetch.Client(ctx), ctx)
		log.Infof(ctx, "after setting http client, token = %s\n", os.Getenv("SPARK_TOKEN"))
	})

  m.Post("/spark", binding.Json(SparkEvent{}), func(sparkEvent SparkEvent, res http.ResponseWriter, req *http.Request, r render.Render) {
		ctx := appengine.NewContext(req)

		//log.Infof(ctx, "Message = %v", sparkEvent)

    message := spark.Message{ID: sparkEvent.Id}
		message.Get()
		log.Infof(ctx, message.Text)

    if strings.HasPrefix(message.Text, "/") {
      s := strings.Split(sparkEvent.Text, " ")

      command := s[0]
      log.Infof(ctx, "command = %s", command)
      if command == "/routes" {

        _, body, _ := gorequest.New().Get("http://galwaybus.herokuapp.com/routes.json").End()
        log.Infof(ctx, "body = %s\n", body)

        var routeMap map[string]BusRoute
        json.Unmarshal([]byte(body), &routeMap)

        text := "Routes:\n\n"
        for _, route := range routeMap {
            text = text + strconv.Itoa(route.Id) + " " + route.LongName + "\n"
        }

        message := spark.Message{
					RoomID: sparkEvent.RoomId,
					Text:   text,
				}
				message.Post()
      }
    }

  })

  http.Handle("/", m)
}
