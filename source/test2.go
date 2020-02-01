package main

import (
    "encoding/json"
    "fmt"
    "io/ioutil"
    "text/template"
    "os"
    "database/sql"
    "strconv"
    "strings"
    
    _ "github.com/mattn/go-sqlite3"
)

func main() {

    // read file
    data, err := ioutil.ReadFile("intents/GetWoll")
    if err != nil {
      fmt.Print(err)
    }


    // define data structure 
    type IntentStruct struct {
      Source string
      Query string
      Say string
    }

    // json data
    var intent IntentStruct

    // unmarshall it
    err = json.Unmarshal(data, &intent)
    if err != nil {
        fmt.Println("error:", err)
    }

    if intent.Source == "static" {
      fmt.Printf("%s\n", intent.Say);
    }
    if intent.Source == "sql" {
      fmt.Printf("%s\n", intent.Say);
      var templateVars int
      templateVars = strings.Count(intent.Say,"{{")

        type Inventory struct {
          Material string
          Count    uint
        }

    
    
        database, _ := sql.Open("sqlite3", "./nraboy.db")
        /*
        statement, _ := database.Prepare("CREATE TABLE IF NOT EXISTS people (id INTEGER PRIMARY KEY, material TEXT)")
        statement.Exec()
        statement, _ = database.Prepare("INSERT INTO people (material) VALUES (?)")
        statement.Exec("wool")
        */
        fmt.Println("hur många vars ", templateVars)
        rows, _ := database.Query("SELECT id, material FROM people LIMIT 2")
        var id int
        var material string
        for rows.Next() {
            rows.Scan(&id, &material)
            fmt.Println(strconv.Itoa(id) + ": " + material)
        }
        
        sweaters := Inventory{material, uint(id)}

        tmpl, err := template.New("test").Parse(intent.Say)
        if err != nil { panic(err) }
        err = tmpl.Execute(os.Stdout, sweaters)
        if err != nil { panic(err) }



    }


 


}
