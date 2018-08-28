package main

import (
	// "bytes"
	"database/sql"
	"fmt"
	xj "github.com/basgys/goxml2json"
	"github.com/buger/jsonparser"
	_ "github.com/lib/pq"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// type GSToRE struct {
// 	Taxonomy string `json:"taxonomy"`
// 	Basename string `json:"basename"`
// 	Sources  []struct {
// 		Mimetype  string   `json:"mimetype"`
// 		Files     []string `json:"files"`
// 		Set       string   `json:"set"`
// 		External  string   `json:"external"`
// 		Extension string   `json:"extension"`
// 	} `json:"sources"`
// 	Dippath     string   `json:"dippath"`
// 	Description string   `json:"description"`
// 	Author      string   `json:"author"`
// 	Apps        []string `json:"apps"`
// 	IsEmbargoed string   `json:"is_embargoed"`
// 	Categories  []struct {
// 		Subtheme  string `json:"subtheme"`
// 		Theme     string `json:"theme"`
// 		Groupname string `json:"groupname"`
// 	} `json:"categories"`
// 	Spatial struct {
// 		Geomtype string `json:"geomtype"`
// 		Epsg     int    `json:"epsg"`
// 		Features int    `json:"features"`
// 		Bbox     string `json:"bbox"`
// 		Records  int    `json:"records"`
// 	} `json:"spatial"`
// 	Metadata struct {
// 		XML      struct{} `json:"xml"`
// 		Upgrade  string   `json:"upgrade"`
// 		Standard string   `json:"standard"`
// 	} `json:"metadata"`
// 	OrigEpsg       int      `json:"orig_epsg"`
// 	Epsg           int      `json:"epsg"`
// 	Totalfiles     int      `json:"totalfiles"`
// 	Firstname      string   `json:"firstname"`
// 	Lastname       string   `json:"lastname"`
// 	Nodeid         string   `json:"nodeid"`
// 	DataoneArchive string   `json:"dataone_archive"`
// 	Active         string   `json:"active"`
// 	Folderlineage  []string `json:"folderlineage"`
// 	Embargo        struct {
// 		ReleaseDate string `json:"release_date"`
// 		Embargoed   bool   `json:"embargoed"`
// 	} `json:"embargo"`
// 	Percentdone int      `json:"percentdone"`
// 	Phone       string   `json:"phone"`
// 	Standards   []string `json:"standards"`
// 	Releasedate string   `json:"releasedate"`
// 	Services    []string `json:"services"`
// 	Formats     []string `json:"formats"`
// }

//VERSION is and exported variable so the handelers can use it.
var VERSION string

//CODE ... I don't remember. Do we even need this?
var CODE string

// CODENAME is like a major version string
var CODENAME string

//altpath is used if you need an alternate path for some web servers.
var altpath string

//dbuser pass and name and secret to be pulled from config
var user string
var password string
var dbname string
var host string
var port int
var psqlInfo string
var formats string
var fileformats []string

func main() {
	VERSION = "0.0"
	CODENAME = "sledge"
	argsWithoutProg := os.Args[1:]
	fmt.Println(argsWithoutProg)
	//If args given...
	if len(argsWithoutProg) > 0 {
		if argsWithoutProg[0] == "install" {
			if _, err := os.Stat("/etc/GSTORE-insert/"); os.IsNotExist(err) {
				pathErr := os.MkdirAll("/etc/GSTORE-insert/", 0777)
				if pathErr != nil {
					fmt.Println(pathErr)
				}
				d1 := []byte("#Log files location\nLogDir = \"/var/log/\"\n\n#DB connection info\nDBUser = \"pguser\"\nDBPass = \"pgpass\"\nDBName = \"gstorepgdb\"")
				err := ioutil.WriteFile("/etc/GSTORE-insert/GSTORE-insert.conf", d1, 0644)
				if err != nil {
					fmt.Println(err)
				}
				os.OpenFile("/var/log/GSTORE-insert.log", os.O_RDONLY|os.O_CREATE, 0666)
			}
		} else {
			fmt.Println("Unknown param")
		}
	} else { //If no args given...

		var configf = ReadConfig() //this is in config.go

		host = configf.DBHost
		port, _ = strconv.Atoi(configf.DBPort)
		user = configf.DBUser
		password = configf.DBPass
		dbname = configf.DBName
		formats = configf.FileFormats
		fileformats = strings.Split(formats, ",")
		psqlInfo = fmt.Sprintf("host=%s port=%d user=%s "+"password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)

		fmt.Println("b")
		//Bulk Insert run every minute..
		ticker := time.NewTicker(time.Second * 10)
		go func() {
			for t := range ticker.C {
				BulkInsert()
				_ = t
			}
		}()

		go forever()
		select {} // block forever

	}
}

//Append is a function for appending slices
func Append(slice []string, items ...string) []string {
	for _, item := range items {
		slice = Extend(slice, item)
	}
	return slice
}
func forever() {
	for {
		// fmt.Printf("%v+\n", time.Now())
		// time.Sleep(time.Second)
		time.Sleep(time.Second * 30)
	}
}

//Extend is an easy wat to grow a slice.
func Extend(slice []string, element string) []string {
	n := len(slice)
	if n == cap(slice) {
		// Slice is full; must grow.
		// We double its size and add 1, so if the size is zero we still grow.
		newSlice := make([]string, len(slice), 2*len(slice)+1)
		copy(newSlice, slice)
		slice = newSlice
	}
	slice = slice[0 : n+1]
	slice[n] = element
	return slice
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func logErr(err error) {
	if err != nil {
		log.Println(err)
	}
}

//BulkInsert check db for DIPS, and if it finds one that has the corect status it runs it.
func BulkInsert() {
	fmt.Printf("%v+\n", time.Now())

	fmt.Println("BulkInsert Running")
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		panic(err)
	}
	defer db.Close()
	err = db.Ping()
	if err != nil {
		panic(err)
	}

	rows, err := db.Query("select json from datasets_in_progress where insert_type='queued'")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	for rows.Next() {
		var json string
		err := rows.Scan(&json)
		if err != nil {
			log.Fatal(err)
		}
		data := []byte(json)
		Folder, valType, Offset, err := jsonparser.Get(data, "baseGeoFolder")
		logErr(err)
		if err == nil {
			// n := bytes.Index(Folder, []byte{0})
			// fmt.Println(string(Folder[:n]))
			folder := string(Folder[:])
			//fmt.Println(folder)
			//fmt.Println(valType)
			//fmt.Println(Offset)
			files, err := filepath.Glob(folder + "*")
			if err != nil {
				fmt.Print(err)
				os.Exit(1)
				//TODO error should be written to db and status changed to error.
			}
			if len(files) == 0 {
				//TODO error to db and change status.
				fmt.Println("No files found in " + folder)
			} else {
				//fmt.Println(files)
				for _, file := range files {
					//bn:=filepath.Base(file)
					//fmt.Println(index)

					

					ext := filepath.Ext(file)
					basename := strings.TrimRight(file, ext)
					if contains(fileformats, ext) {

						if ext == ".tif" || ext == ".tiff" {
							SourcesFiles:=[]
							demcheck, err := filepath.Glob(basename + ".dem")
							logErr(err)
							//if demcheck>0){
							xmls, err := filepath.Glob(basename + "*.xml")
							//fmt.Println(xmls)
							  if err != nil {
								//TODO Handle problem listing folders
								fmt.Println(err)

							} else {
								//xmlchoice := ""
								if len(xmls) > 0 {
									for _, xmlfilepath := range xmls {
										//fmt.Println(xmlfile)
										//fmt.Println(xmlindex)
										if strings.Contains(strings.ToLower(xmlfilepath), "fgdc") {
											// xmlchoice = xmlfilepath
											//fmt.Println(xmlchoice)
											//	fmt.Println(file)
											xmlfile, err := os.Open(xmlfilepath)
											if err != nil {
												log.Fatal(err)
											}
											// xml := strings.NewReader(xmlfile)
											xmljson, err := xj.Convert(xmlfile)
											if err != nil {
												panic("That's embarrassing...")
											}

											//	fmt.Println(xmljson.String())
											//GSToREJSON := `{"taxonomy":"vector","basename":"` + basename + `","sources":[{"mimetype":"application/x-zip-compressed","files":["/geodata/epscor1_CoverageCatalog/ROOT/Structures/NM_Address_20180613.zip"],"set":"original","external":"False","extension":"zip"}],"dippath":"RGIS/Structures","description":"New Mexico Structure Points 6-13-2018","author":"New Mexico Department of Finance and Administration","apps":["rgis"],"is_embargoed":"False","categories":[{"subtheme":"General","theme":"Land Use/Land Cover","groupname":"New Mexico"}],"spatial":{"geomtype":"POINT","epsg":4326,"features":939605,"bbox":"-109.309196418,31.3202520319,-103.051593326,37.0455301703","records":939605},"email":"gstore@edac.unm.ed","metadata":{"xml":` + xmljson.String() + `,"upgrade":"true","standard":"FGDC-STD-001-1998"},"orig_epsg":4326,"epsg":4326,"totalfiles":100,"firstname":"GSToRE","lastname":"Admin","nodeid":"9bd55a2e-474b-4167-b92f-008df1c0adb5","dataone_archive":"False","active":"True","folderlineage":["d2b7bf1e-d466-11e7-8325-cf521938ae66","9bd55a2e-474b-4167-b92f-008df1c0adb5"],"embargo":{"release_date":"2014-05-02","embargoed":false},"percentdone":0,"phone":"505-277-3622","standards":["FGDC-STD-001-1998","ISO-19115:2003"],"releasedate":"2018-08-27","services":["wms","wcs"],"formats":["zip"]}`
											GSToREJSON := `{
												"taxonomy": "geoimage",
												"basename": "` + basename + `",
												"sources": [{
													"mimetype": "application/x-zip-compressed",
													"files": ["/geodata/epscor1_CoverageCatalog/ROOT/Structures/NM_Address_20180613.zip"],
													"set": "original",
													"external": "False",
													"extension": "zip"
												}],
												"dippath": "RGIS/Structures",
												"description": "New Mexico Structure Points 6-13-2018",
												"author": "New Mexico Department of Finance and Administration",
												"apps": ["rgis"],
												"is_embargoed": "False",
												"categories": [{
													"subtheme": "General",
													"theme": "Land Use/Land Cover",
													"groupname": "New Mexico"
												}],
												"spatial": {
													"geomtype": "POINT",
													"epsg": 4326,
													"features": 939605,
													"bbox": "-109.309196418,31.3202520319,-103.051593326,37.0455301703",
													"records": 939605
												},
												"email": "gstore@edac.unm.ed",
												"metadata": {
													"xml": ` + xmljson.String() + `,
												"orig_epsg": 4326,
												"epsg": 4326,
												"totalfiles": 100,
												"firstname": "GSToRE",
												"lastname": "Admin",
												"nodeid": "9bd55a2e-474b-4167-b92f-008df1c0adb5",
												"dataone_archive": "False",
												"active": "True",
												"folderlineage": [
													"d2b7bf1e-d466-11e7-8325-cf521938ae66",
													"9bd55a2e-474b-4167-b92f-008df1c0adb5"
												],
												"embargo": {
													"release_date": "2014-05-02",
													"embargoed": false
												},
												"percentdone": 0,
												"phone": "505-277-3622",
												"standards": [
													"FGDC-STD-001-1998",
													"ISO-19115:2003"
												],
												"releasedate": "2018-08-27",
												"services": [
													"wms",
													"wCs"
												],
												"formats": ["zip"]
											}`

											fmt.Println(GSToREJSON)

										}
									}
								} else {
									//TODO Handle this error
									fmt.Println("NO XML FOUND!!")
								}
							}
						}
					}

				}
			}
		}

	}

}
