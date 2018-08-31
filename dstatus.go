package router

import (
	"bytes"
	"html/template"
	"log"
	"math"
	"net/http"
	"strings"
	"sync"
	"time"
)

type AdditionalPrintF func() string

type Dstatus struct {
	tmpl       *template.Template
	data       DstatusData
	actions    DstatusActions
	Additional AdditionalPrintF
	IsDetailed bool
}

type DstatusData struct {
	Name       string
	Port       string
	Country    string
	CommitId   string
	CommitName string
	StartTime  time.Time
}

func (dstatus *Dstatus) init() {
	dstatus.tmpl = template.New("dstatus")
	if dstatus.data.StartTime.IsZero() {
		dstatus.data.StartTime = time.Now()
	}
	var err error
	//dstatus.tmpl, err = template.ParseFiles(path.Dir(filename) + "/template/dstatus.html")
	dstatus.tmpl, err = template.ParseFiles("./template/dstatus.html")
	if err != nil {
		log.Fatal(err)
	}
	dstatus.actions.m = make(DstatusActionMap)
}

func (dstatus *Dstatus) Show(writer http.ResponseWriter, request *http.Request, uriParts []string) {
	//"↳"
	var tpl bytes.Buffer

	additionalData := ""
	if dstatus.Additional != nil {
		additionalData = dstatus.Additional()
	}
	dstatus.actions.RLock()
	defer dstatus.actions.RUnlock()
	data := struct {
		Name           string
		Ip             string
		Port           string
		Country        string
		Uri            string
		StartTime      time.Time
		Duration       time.Duration
		CommitId       string
		CommitName     string
		Map            DstatusActionMap
		AdditionalData template.HTML
	}{
		Ip:  strings.Split(request.RemoteAddr, ":")[0],
		Uri: request.URL.String(),

		Name:       dstatus.data.Name,
		Port:       dstatus.data.Port,
		StartTime:  dstatus.data.StartTime,
		Duration:   time.Since(dstatus.data.StartTime),
		CommitId:   dstatus.data.CommitId,
		CommitName: dstatus.data.CommitName,
		Map:        dstatus.actions.m,

		Country:        "Russia",
		AdditionalData: template.HTML(additionalData),
	}

	if err := dstatus.tmpl.Execute(&tpl, data); err != nil {
		log.Fatal(err)
	}
	writer.Write([]byte(tpl.String()))
}

func (dstatus *Dstatus) HandleWrapper(f HandlerFunc) HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request, uriParts []string) {
		start := time.Now()
		f(w, r, uriParts)
		for idx := range uriParts {
			action := strings.Join(uriParts[0:idx], "↳")
			if len(action) > 0 {

				dstatus.actions.update(action, time.Since(start).Seconds())
			}
			if !dstatus.IsDetailed && idx > 0 {
				break
			}
		}
	}
}
func NewDstatus(data DstatusData) *Dstatus {
	dstatus := Dstatus{data: data}
	dstatus.init()
	return &dstatus
}

type DstatusActionData struct {
	Count        uint64
	PerSec       float64
	MinExecTime  float64
	MaxExecTime  float64
	ExecTime     float64
	AvgExecTime  float64
	LongRequests uint64
	Start        time.Time
	Last         time.Time
}

type DstatusActionMap map[string]*DstatusActionData
type DstatusActions struct {
	sync.RWMutex
	m DstatusActionMap
}

func (actions *DstatusActions) update(action string, d float64) {
	actions.Lock()
	defer actions.Unlock()
	var longRequests uint64
	if d > 1 {
		longRequests = 1
	}
	now := time.Now()
	if val, ok := actions.m[action]; ok {
		val.Count += 1
		val.MinExecTime = math.Min(val.MinExecTime, d)
		val.MaxExecTime = math.Max(val.MaxExecTime, d)
		val.ExecTime += d
		val.LongRequests += longRequests
		val.AvgExecTime = val.ExecTime / float64(val.Count)
		val.PerSec = float64(val.Count) / (time.Since(val.Start).Seconds())
		val.Last = now
	} else {
		actions.m[action] = &DstatusActionData{
			Count:        1,
			MinExecTime:  d,
			MaxExecTime:  d,
			ExecTime:     d,
			LongRequests: longRequests,
			AvgExecTime:  d,
			Start:        now,
			Last:         now,
		}
	}
}
