# godbi

_godbi_ adds a set of high-level functions to the generic SQL handle in GO. Check *godoc* for definitions:
[![GoDoc](https://godoc.org/github.com/genelet/godbi?status.svg)](https://godoc.org/github.com/genelet/godbi)

There are three levels of usages:

- _Basic_: operating on raw SQL statements and receive data as *[]map[string]interface{}*
- _Model_: fulfilling CRUD actions on a table.
- _Schema_: fulfilling RESTful and GraphQL actions on multiple tables.

_godbi_ runs CRUD and RESTful tasks easily and gracefully.
The package is fully tested in MySQL and PostgreSQL.

<br /><br />

### Installation

> $ go get -u github.com/genelet/godbi
<!-- go mod init github.com/genelet/godbi -->

### Termilogy

The names of arguments passed in functions or methods in this package are defined as follows, if not specifically explained:
Name | Type | IN/OUT | Where | Meaning
---- | ---- | ------ | ----- | -------
*args* | `...interface{}` | IN | `DBI` | slice of values, possibly empty
*ARGS* | `map[string]interface{}` | IN | `Model`,`Graph` | input data
*extra* | `...map[string]interface{}` | IN | `Model`,`Schema` | for WHERE constraints, possibly empty
*lists* | `[]map[string]interface{}` | OUT | all | data received as a slice of maps.

<br /><br />

## Chapter 1. BASIC USAGE

### 1.1  Type _DBI_

The `DBI` type simply embeds the standard SQL handle.

```go
package godbi

type DBI struct {
    *sql.DB          
    LastID    int64  // saves the last inserted id
}

```

#### 1.1.1) Create a new handle

```go
dbi := &DBI{DB: the_standard_sql_handle}
```

#### 1.1.2) Example

In this example, we create a MySQL handle using credentials in the environment; then create a new table _letters_ with 3 rows. We query the data using `SelectSQL` and put the result into `lists`.

<details>
    <summary>Click for Sample 1</summary>
    <p>

```go
package main

import (
    "os"
    "log"
    "database/sql"
    "github.com/genelet/godbi"
    _ "github.com/go-sql-driver/mysql"
)

func main() {
    dbUser := os.Getenv("DBUSER")
    dbPass := os.Getenv("DBPASS")
    dbName := os.Getenv("DBNAME")
    db, err := sql.Open("mysql", dbUser + ":" + dbPass + "@/" + dbName)
    if err != nil { panic(err) }
    defer db.Close()

    dbi := &godbi.DBI{DB:db}

    // create a new table and insert some data using ExecSQL
    //
    _, err = db.Exec(`DROP TABLE IF EXISTS letters`)
    if err != nil { panic(err) }
    _, err = db.Exec(`CREATE TABLE letters (
        id int auto_increment primary key,
        x varchar(1))`)
    if err != nil { panic(err) }
    _, err = db.Exec(`insert into letters (x) values ('m')`)
    if err != nil { panic(err) }
    _, err = db.Exec(`insert into letters (x) values ('n')`)
    if err != nil { panic(err) }
    _, err = db.Exec(`insert into letters (x) values ('p')`)
    if err != nil { panic(err) }

    // select all data from the table
    lists := make([]map[string]interface{}, 0)
    err = dbi.SelectSQL(&lists, "SELECT id, x FROM letters")
    if err != nil { panic(err) }

    log.Printf("%v", lists)

    dbi.Exec(`DROP TABLE IF EXISTS letters`)
    db.Close()
}
```

Running this example will result in something like

```bash
[map[id:1 x:m] map[id:2 x:n] map[id:3 x:p]]
```

</p>
</details>

<br /><br />

### 1.2  Execution `DoSQL`

```go
func (*DBI) DoSQL  (query string, args ...interface{}) error
```

The same as DB's `Exec`, except it saves the possible last id and returns error only. 

Note that most of the functions in this package return error, which should be checked to assert the execution status.

<br /><br />

### 1.3   _SELECT_ Queries

#### 1.3.1)  `SelectSQL`

```go
func (*DBI) Select(lists *[]map[string]interface{}, query string, args ...interface{}) error
```

Run the *SELECT*-type query and save the result in `lists`. The data types of the column are determined dynamically by the generic SQL handle.

<details>
    <summary>Click for example</summary>
    <p>

```go
lists := make([]map[string]interface{})
err = dbi.Select(&lists,
    `SELECT ts, id, name, len, flag, fv FROM mytable WHERE id=?`, 1234)
```

will select all rows with *id=1234*.

```json
    {"ts":"2019-12-15 01:01:01", "id":1234, "name":"company", "len":30, "flag":true, "fv":789.123},
    ....
```

</p>
</details>


#### 1.3.2) `SelectSQL`

```go
func (*DBI) SelectSQL(lists *[]map[string]interface{}, labels []interface{}, query string, args ...interface{}) error
```

It differs from the above `Select` by specifying map's key names, and optionally their data types, in `labels`, depending if interface is `string` or `[2]string`.

The following example assigns key names TS, id, Name, Length, Flag and fv, and
data types _string_, _int_, _string_, _int8_, _bool_ and _float32_ to the corresponding columns:

<details>
    <summary>Click for example</summary>
    <p>

```go
lists := make([]map[string]interface{})
err = dbi.querySQLLabel(&lists, 
    `SELECT ts, id, name, len, flag, fv FROM mytable WHERE id=?`,
	[]interface{}{[2]string{"TS","string"], [2]string{"id","int"], [2]string{"Name","string"], [2]string{"Length","int8"], [2]string{"Flag","bool"], [2]string{"fv","float32"]},
    1234)
```

```json
    {"TS":"2019-12-15 01:01:01", "id":1234, "Name":"company", "Length":30, "Flag":true, "fv":789.123},
```

</p>
</details>


<br /><br />

### 1.4  _GetSQL_ Single Row

If only one row will be returned from a query, this function returns the data as a map.


```go
func (*DBI) GetSQL(res map[string]interface{}, query string, labels []interface{}, args ...interface{}) error
```

<br /><br />

## Chapter 2. MODEL USAGE

In this type of usage, we define any customized `action` using interface `Capability`
and run it against `table`. The following CRUD actions are pre-defined:
C | R | U | D | P
---- | ---- | ---- | ---- | ----
create a new row | read all rows, or read one row | update a row | delete a row | insert or update

They correspond to the REST web methods as follows:

<details>
    <summary>Click for RESTful vs CRUD</summary>
    <p>

HTTP METHOD | Web URL | CRUD | Function in godbi
----------- | ------- | ---- | -----------------
GET         | webHandler | R All | Topics
GET         | webHandler/ID | R One | Edit
POST        | webHandler | C | Insert
PUT         | webHandler | U | Update
PATCH       | webHandler | P | Insupd
DELETE      | webHandler | D | Delete

</p>
</details>

<br /><br />

### 2.1  Type *Table*

`Table` describes simple table properties.

```go
type Table struct {
    CurrentTable   string    `json:"current_table,omitempty"`   // the current table name
    Pks            []string  `json:"pks,omitempty"`    // optional, the PK 
    IDAuto         string    `json:"id_auto,omitempty"` // this table has an auto id
	Fks            []string  `json:"fks,omitempty"`    // optional, the PK
}
```

### 2.2  Type *Action*

`Action` defines an action on the table.

```go
type Action struct {
    Must      []string    `json:"must,omitempty"
    Nextpages []*Page     `json:"nextpages,omitempty"
    Appendix  interface{} `json:"appendix,omitempty"
}
```
Every action should implement function `RunActionContext` using the `Capability` interface:
```go
type Capability interface {
    RunActionContext(context.Context, *sql.DB, *Table, map[string]interface{}, ...map[string]interface{}) ([]map[string]interface{}, []*Page, error)
}
```

#### 2.2.1) *Insert* 

```go
type Insert struct {
    Action
    Columns    []string      `json:"columns,omitempty" hcl:"columns,optional"`
}
```

This action inserts one row into the table. The row data is expressed as `map[string]interface{}`.

#### 2.2.2) *Update* 

### 2.3  *Nextpages* for Follow-up Actions

As in GraphQL, *godbi* allows an action to trigger multiple actions on other models. *Nextpages* 
which is a slice of *Page*, defines such follow-up actions:

```go
type Page struct {
    Model      string            `json:"model"`                 // name of the next model to call  
    Action     string            `json:"action"`                // action name of the next model
    RelateItem map[string]string `json:"relate_item,omitempty"` // column name mapped to that of the next model
    Extra      map[string]string `json:"extra,omitempty"`      // manually assign these constraints
}
```

Here is a use case. There are two tables, one for family and the other for children, corresponding to models `ta` and `tb` respectively.
We *query* the family name in `ta`, and want to show all children under the family names. Technically, it means we need to run
a `Topics` action of `tb` on each row in the family data, constrained by the association of family's ID in both the tables. So `Nextpages`
of `ta` will look like:

<details>
    <summary>Click to show the JSON string</summary>
    <p>

```json
{
    "edit" : [
        {"model":"tb", "action":"topics", "relate_item":{"id":"id"}}
    ],
    "topics" : [
        {"model":"tb", "action":"topics", "relate_item":{"id":"id"}}
    ]
}
```

</p>
</details>

Parsing the JSON will build up the `Nextpages` as `map[string][]*Page`.

<br /><br />

### 2.3  Type *Model*

`Model` contains both the table and all actions on the table. 
```go
type Model struct {
	Table
	Actions map[string]interface{} `json:"actions,omitempty" hcl:"actions,optional"`
```

The best way to build up a new `Model` is to parse it from a JSON string or file.

<details>
    <summary>Click to show a full Model example</summary>
    <p>

```go
```

</p>
</details>

#### 2.3.1 `NewModelJson`

```go
func NewModelJson(bs []byte, custom ...map[string]Capability) (*Model, error)
```
where `bs` is the JSON string in bytes, custom is the map between customized action names and their `Capabilities`. Note that
if you are using only the CRUD actions, there is no need to pass `custom`.

#### 2.3.2 `NewModelJsonFile`

```go
func NewModelJsonFile(fn string, custom ...map[string]Capability) (*Model, error)
```
Parse `Model` from disk file `fn`.

#### 2.3.3 `Assertion`

If run own JSON parsing function, we should run `Assertion` to assert the JSON structure to 
the right `Action` types:
```go
func (self *Model) Assertion(custom ...map[string]Capability) error
```

<br /><br />

## 3. `Graph` Usage

### 3.1 Optional Constraints

For all RESTful methods of *Model*, we have option to put a data structure, named `extra` and of type `url.Values`, to constrain the *WHERE* statement. Currently we have supported 3 cases:

<details>
    <summary>Click to show *extra*</summary>
    <p>

key in `extra` | meaning
--------------------------- | -------
key has only one value | an EQUAL constraint
key has multiple values | an IN constraint
key is named *_gsql* | a raw SQL statement
among multiple keys | AND conditions.

</p>
</details>

#### 2.2.4) Returning Data

After we have run an action on the model, we can retrieve data using

```go
(*Model) GetLists()
```

The closure associated with the action name can be get back:

```go
(*Model) GetAction(name string) func(...url.Values) error
```

<br /><br />

### 2.3) Methods of *Model*

#### 2.3.1) For Http METHOD: GET (read all)

```go
func (*Model) Topics(extra ...url.Values) error
```

#### 2.3.2) For Http METHOD: GET (read one)

```go
func (*Model) Edit(extra ...url.Values) error
```

#### 2.3.3) For Http METHOD: POST (create)

```go
func (*Model) Insert(extra ...url.Values) error
```

It inserts a new row using the input data. If `extra` is passed in, it will override the input data.

#### 2.3.4) Http METHOD: PUT (update)

```go
func (*Model) Update(extra ...url.Values) error
```

It updates a row using the input data, constrained by `extra`.

#### 2.3.5) Http METHOD: PATCH (insupd)

```go
func (*Model) Insupd(extra ...url.Values) error
```

It inserts or updates a row using the input data, constrained optionally by `extra`.

#### 2.3.6) Http METHOD: DELETE

```go
func (*Model) Delete(extra ...url.Values) error
```

It rows constrained by `extra`. For this function, the input data will NOT be used.

#### 2.3.7）Example

<details>
    <summary>Click for example to run RESTful actions</summary>
    <p>

```go
package main

import (
    "log"
    "os"
    "net/url"
    "database/sql"
    "github.com/genelet/godbi"
    _ "github.com/go-sql-driver/mysql"
)

func main() {
    dbUser := os.Getenv("DBUSER")
    dbPass := os.Getenv("DBPASS")
    dbName := os.Getenv("DBNAME")
    db, err := sql.Open("mysql", dbUser + ":" + dbPass + "@/" + dbName)
    if err != nil { panic(err) }
    defer db.Close()

    model := new(godbi.Model)
    model.CurrentTable = "testing"
    model.Sortby        ="sortby"
    model.Sortreverse   ="sortreverse"
    model.Pageno        ="pageno"
    model.Rowcount      ="rowcount"
    model.Totalno       ="totalno"
    model.Maxpageno     ="max_pageno"
    model.Fields        ="fields"
    model.Empties       ="empties"

    db.Exec(`DROP TABLE IF EXISTS testing`)
    db.Exec(`CREATE TABLE testing (id int auto_increment, x varchar(255), y varchar(255), primary key (id))`)

    args := make(url.Values)
    model.SetDB(db)
    model.SetArgs(args)

    model.CurrentKey    = "id"
    model.CurrentIDAuto = "id"
    model.InsertPars    = []string{     "x","y"}
    model.TopicsPars    = []string{"id","x","y"}
    model.UpdatePars    = []string{"id","x","y"}
    model.EditPars      = []string{"id","x","y"}

    args["x"] = []string{"a"}
    args["y"] = []string{"b"}
    if err := model.Insert(); err != nil { panic(err) }
    log.Println(model.LastID)

    args["x"] = []string{"c"}
    args["y"] = []string{"d"}
    if err := model.Insert(); err != nil { panic(err) }
    log.Println(model.LastID)

    if err := model.Topics(); err != nil { panic(err) }
    log.Println(model.GetLists())

    args.Set("id","2")
    args["x"] = []string{"c"}
    args["y"] = []string{"z"}
    if err := model.Update(); err != nil { panic(err) }
    if err := model.Edit(); err != nil { panic(err) }
    log.Println(model.GetLists())

    os.Exit(0)
}
```

Running the program will result in

```bash
1
2
[map[id:1 x:a y:b] map[id:2 x:c y:d]]
[map[id:2 x:c y:z]]
```

</p>
</details>

<br /><br />

### 2.4 Action by Name, `GetAction`

Besides running action directly by calling its method, we can run it alternatively by calling its string name. This is important in web application where the server is open to many different user actions and response particular one dynamically according to user query.

To achieve this, build up a map between action name and function (a closure). To get the closure back, use:

```go
// defining the action map 
actions := make(map[string]func(...url.Values) error)
actions["topics"] = func(extra ...url.Values) error { return model.Topics(extra...) }
model.Actions = actions
// later
if acting := model.GetAction("topics"); acting != nil {
    err := acting(extra...)
}
```

which is equivalent to `err := model.Topics(extra...)`.

<br /><br />

## Chapter 3  ADVANCED USAGE

### 3.1 Type `Schema`

Because models are allowed to interact with each other, we introduce type `Schema` which handles the whole database schema at once:

```go
type Schema struct {
    // private fields
    Models  map[string]Navigate
}
```

where keys in the map are assigned model name strings.

#### 3.1.1) Create Schema, `NewSchema`

Create a new schema instance by passing the string to `Model` map.

```go
NewSchema(models map[string]Navigate) *Schema
```

#### 3.1.2) Assign DB, `SetDB`

After a new schema is created, we assign it a database handle:

```go
(* Schema) SetDB(db *sql.DB)
```

#### 3.1.2) Get Model by Name, `GetNavigate`

We can get a model by name

```go
(*Schema) GetNavigate(args url.Values) Navigate
```

Here we pass in the input data as well, so the interface can be cast to the model with input and database handle embedded.

### 3.2) Run a RESTful action

`Schema` implement the `Run` method which is ideal for RESTful requests.

```go
func (*Schema) Run(model, action string, args url.Values, extra ...url.Values) ([]map[string]interface{}, error)
```

Here we pass in the string names of model and action, the input data `args`, the database handle `db`, and optional `extra` constraints. It runs the action and **all next pages defined inside the schema**. The return is the data and optional error.

Here is a full example that covers most information in Chapter 3.

<details>
    <summary>Click for RESTful example using Schema</summary>
    <p>

```go
package main

import (
    "log"
    "os"
    "net/url"
    "database/sql"
    "github.com/genelet/godbi"
    _ "github.com/go-sql-driver/mysql"
)

func main() {
    dbUser := os.Getenv("DBUSER")
    dbPass := os.Getenv("DBPASS")
    dbName := os.Getenv("DBNAME")
    db, err := sql.Open("mysql", dbUser + ":" + dbPass + "@/" + dbName)
    if err != nil { panic(err) }
    defer db.Close()

    db.Exec(`drop table if exists test_a`)
    db.Exec(`CREATE TABLE test_a (id int auto_increment not null primary key,
        x varchar(8), y varchar(8), z varchar(8))`)
    db.Exec(`drop table if exists test_b`)
    db.Exec(`CREATE TABLE test_b (tid int auto_increment not null primary key,
        child varchar(8), id int)`)

    ta, err := godbi.NewModel("test_a.json")
    if err != nil { panic(err) }
    tb, err := godbi.NewModel("test_b.json")
    if err != nil { panic(err) }

    // create action map for ta, the value of map is closure
    //
    action_ta := make(map[string]func(...url.Values)error)
    action_ta["topics"] = func(args ...url.Values) error { return ta.Topics(args...) }
    action_ta["insert"] = func(args ...url.Values) error { return ta.Insert(args...) }
    action_ta["insupd"] = func(args ...url.Values) error { return ta.Insupd(args...) }
    action_ta["delete"] = func(args ...url.Values) error { return ta.Delete(args...) }
    action_ta["edit"]   = func(args ...url.Values) error { return ta.Edit(args...) }
    ta.SetActions(action_ta)

    // create action map for ta, the value of map is closure
    //
    action_tb := make(map[string]func(...url.Values)error)
    action_tb["topics"] = func(args ...url.Values) error { return tb.Topics(args...) }
    action_tb["insert"] = func(args ...url.Values) error { return tb.Insert(args...) }
    action_tb["update"] = func(args ...url.Values) error { return tb.Update(args...) }
    action_tb["delete"] = func(args ...url.Values) error { return tb.Delete(args...) }
    action_tb["edit"]   = func(args ...url.Values) error { return tb.Edit(args...) }
    tb.SetActions(action_tb)

    schema := &godbi.Schema{db, map[string]godbi.Navigate{"ta":ta, "tb":tb}}

    methods := map[string]string{"GET":"topics", "GET_one":"edit", "POST":"insert", "PATCH":"insupd", "PUT":"update", "DELETE":"delete"}

    var lists []map[string]interface{}
    // the 1st web requests is assumed to create id=1 to the test_a and test_b tables:
    //
    args := url.Values{"x":[]string{"a1234567"},"y":[]string{"b1234567"},"z":[]string{"temp"}, "child":[]string{"john"}}
    if lists, err = schema.Run("ta", methods["PATCH"], args); err != nil { panic(err) }

    // the 2nd request just updates, because [x,y] is defined to the unique in ta.
    // but create a new record to tb for id=1, since insupd triggers insert in tb
    //
    args = url.Values{"x":[]string{"a1234567"},"y":[]string{"b1234567"},"z":[]string{"zzzzz"}, "child":[]string{"sam"}}
    if lists, err = schema.Run("ta", methods["PATCH"], args); err != nil { panic(err) }

    // the 3rd request creates id=2
    //
    args = url.Values{"x":[]string{"c1234567"},"y":[]string{"d1234567"},"z":[]string{"e1234"},"child":[]string{"mary"}}
    if lists, err = schema.Run("ta", methods["POST"], args); err != nil { panic(err) }

    // the 4th request creates id=3
    //
    args = url.Values{"x":[]string{"e1234567"},"y":[]string{"f1234567"},"z":[]string{"e1234"},"child":[]string{"marcus"}}
    if lists, err = schema.Run("ta", methods["POST"], args); err != nil { panic(err) }

    // GET all
    args = url.Values{}
    if lists, err = schema.Run("ta", methods["GET"], args); err != nil { panic(err) }
    log.Printf("%v", lists)

    // GET one
    args = url.Values{"id":[]string{"1"}}
    if lists, err = schema.Run("ta", methods["GET_one"], args); err != nil { panic(err) }
    log.Printf("%v", lists)

    // DELETE
    extra := url.Values{"id":[]string{"1"}}
    if lists, err = schema.Run("tb", methods["DELETE"], url.Values{}, extra); err != nil { panic(err) }
    if lists, err = schema.Run("ta", methods["DELETE"], url.Values{}, extra); err != nil { panic(err) }

    // GET all
    args = url.Values{}
    if lists, err = schema.Run("ta", methods["GET"], args); err != nil { panic(err) }
    log.Printf("%v", lists)

    os.Exit(0)
}
```

Running it will result in:

```bash
[map[id:1 tb_topics:[map[child:john id:1 tid:1] map[child:sam id:1 tid:2]] x:a1234567 y:b1234567 z:zzzzz] map[id:2 tb_topics:[map[child:mary id:2 tid:3]] x:c1234567 y:d1234567 z:e1234] map[id:3 tb_topics:[map[child:marcus id:3 tid:4]] x:e1234567 y:f1234567 z:e1234]]
[map[id:1 tb_topics:[map[child:john id:1 tid:1] map[child:sam id:1 tid:2]] x:a1234567 y:b1234567 z:zzzzz]]
[map[id:2 tb_topics:[map[child:mary id:2 tid:3]] x:c1234567 y:d1234567 z:e1234] map[id:3 tb_topics:[map[child:marcus id:3 tid:4]] x:e1234567 y:f1234567 z:e1234]]
```

</p>
</details>
