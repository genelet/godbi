# godbi

_godbi_ adds a set of high-level functions to the generic SQL handle in GO. Check *godoc* for definitions:
[![GoDoc](https://godoc.org/github.com/genelet/godbi?status.svg)](https://godoc.org/github.com/genelet/godbi)

There are three levels of usages:

- _Basic_: operating on raw SQL statements and receive data as *[]map[string]interface{}*
- _Model_: fulfilling CRUD actions on a table.
- _Graph_: fulfilling GraphQL actions on multiple tables.

_godbi_ runs CRUD and RESTful tasks easily and gracefully.
The package is fully tested in MySQL and PostgreSQL.

<br /><br />

### Installation

> $ go get -u github.com/genelet/godbi
<!-- go mod init github.com/genelet/godbi -->

### Termilogy

The names of arguments passed in functions or methods are defined as follows, if not specifically explained:
Name | Type | IN/OUT | Meaning
---- | ---- | ------ | -------
*args* | `...interface{}` | IN | function's arguments, possibly empty
*ARGS* | `map[string]interface{}` | IN | input data
*extra* | `...map[string]interface{}` | IN | _WHERE_ constraints, possibly empty
*lists* | `[]map[string]interface{}` | OUT | receiving data as a slice of maps.

<br />

Most functions in this package return error, which can be checked to assert the execution status.
<br /><br />

## Chapter 1. BASIC USAGE

### 1.1  _DBI_

The `DBI` type simply embeds the standard SQL handle.

```go
package godbi

type DBI struct {
    *sql.DB          
    LastID    int64  // saves the last inserted id
}

```

To create a new handle

```go
dbi := &DBI{DB: the_standard_sql_handle}
```

<br /><br />

### 1.2  `DoSQL`

```go
func (*DBI) DoSQL(query string, args ...interface{}) error
```

The same as DB's `Exec`, except it returns error only. 

<br /><br />

### 1.3  `TxSQL`

```go
func (*DBI) TxSQL(query string, args ...interface{}) error
```

The same as `DoSQL`, but use transaction. 

<br /><br />

### 1.4   _Select_

#### 1.4.1)  `Select`

```go
func (*DBI) Select(lists *[]map[string]interface{}, query string, args ...interface{}) error
```

It runs a *Select*-type query and saves the result into `lists`. The data types of the column are determined dynamically by the generic SQL handle.

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


#### 1.4.2) `SelectSQL`

```go
func (*DBI) SelectSQL(lists *[]map[string]interface{}, labels []interface{}, query string, args ...interface{}) error
```

It selects by specifying map keys, and optionally their data types in `labels`, depending on if it is `string` or `[2]string`.

The following example assigns key names _TS_, _id_, _Name_, _Length_, _Flag_ and _fv_, of data types _string_, _int_, _string_, _int8_, _bool_ and _float32_ to the corresponding columns:

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

### 1.5  _GetSQL_

If there is only one data row returned, this function returns it as a map.


```go
func (*DBI) GetSQL(res map[string]interface{}, query string, labels []interface{}, args ...interface{}) error
```

### 1.6) Example

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

## Chapter 2. MODEL USAGE

In this usage, we start by define an `action` on database table.
The following _CRUD_ actions, and associated _REST_ methods, are pre-defined in our package:

HTTP METHOD | Web URL | CRUD | Function Name | Meaning
----------- | ------- | ---- | ------------- | ----------------
_LIST_        | webHandler | R All | _Topics_ | read all rows
_GET_         | webHandler/ID | R One | _Edit_ | read row by ID
_POST_        | webHandler | C | _Insert_ | create a new row
_PUT_         | webHandler | U | _Update_ | update a row
_PATCH_       | webHandler | P | _Insupd_ | update or insert
_DELETE_      | webHandler | D | _Delete_ | delete a row

<br /><br />

### 2.1  *Table*

_Table_ describes a database table.

```go
type Table struct {
    CurrentTable  string    `json:"current_table,omitempty"`  // the table name
    Pks           []string  `json:"pks,omitempty"`     // optional, the PK 
    IDAuto        string    `json:"id_auto,omitempty"` // table's auto id
    Fks           []string  `json:"fks,omitempty"`     // optional, for the FK
}
```

where _CurrentTable_ is the table name; _Pks_ the primary key (which could be a combination of multiple columns); _IDAuto_ the column whose values is a series number and _Fks_ the foreign key information.

_Fks_ does not need to be a native foreign key defined in relational database, but a relationship between two tables. Currently, we only support foreign
tables which use a single column as its primary key. _Fks_ is defined:

index | meaning
----- | -------------------------
0 | the foreign table name
1 | the primary key name in the foreign table
2 | the signature name of foreign table's primary key
3 | the corresponding column of foreign table's PK in the current table
4 | the signature name of the corresponding column

<br />

### 2.2  *Action*

`Action` defines an action on table. It should implement function `RunActionContext` using the `Capability` interface:

```go
type Action struct {
    Must      []string    `json:"must,omitempty"
    Nextpages []*Edge     `json:"nextpages,omitempty"
    Appendix  interface{} `json:"appendix,omitempty"
}

type Capability interface {
    RunActionContext(context.Context, *sql.DB, *Table, map[string]interface{}, ...map[string]interface{}) ([]map[string]interface{}, []*Edge, error)
}
```
where _Must_ defines non-empty columns; _Nextpages_ a list of other actions, expressed in _Edge_, to follow after the current action is complete; and _Appendix_ stores optional data. For _Edge_, see below.

In _RunActionContext_, the input data _ARGS_ is expressed as _map[string]interface{}_.

<br />

#### 2.2.1) *Insert* 

```go
type Insert struct {
    Action
    Columns    []string      `json:"columns,omitempty" hcl:"columns,optional"`
}
```

It inserts one row into _Columns_ of the table. Row's data is expressed as _map[string]interface{}_ in _RunActionContext_.

#### 2.2.2) *Update* 

```go
ype Update struct {
    Action
    Columns    []string      `json:"columns,omitempty" hcl:"columns,optional"`
    Empties    []string      `json:"empties,omitempty" hcl:"empties,optional"`
}
```

It updates a row using the primary key. _Empties_ is a list of columns whose values should be forced to be empty or null, when there is no input data.

#### 2.2.3) *Insupd* 

```go
ype Insupd struct {
    Action
    Columns    []string      `json:"columns,omitempty" hcl:"columns,optional"`
    Uniques    []string      `json:"uniques,omitempty" hcl:"uniques,optional"`
}
```

It updates a row by checking the data in the unique columns _Uniques_. If not exists, it will insert a new row instead of updating.

#### 2.2.4) *Edit* 

```go
ype Edit struct {
    Action
    Joins    []*Join             `json:"joins,omitempty" hcl:"join,block"`
    Rename   map[string][]string `json:"rename" hcl:"rename"`
    FIELDS   string              `json:"fields,omitempty" hcl:"fields"`
}
```

It reads a row or JOINed-table row using the primary key. See below for _Join_. The selected columns are keys in `Rename`. The value of such a key has two elements. The first one is the renamed label and the second the GO data type like _int_, _int64_ and _string_ etc.

If we need only a part of the row, not the whole row, we can use _FIELDS_ which is a key in input data _ARGS_. Its value is a shorter list of columns, separated by comma. For example, we have defined _FIELDS ="fields"_. In order to return just the user id and username, our input data should have _ARGS["fields"] = "user_id,username"_.

#### 2.2.5) *Topics* 

```go
    Action
    Joins       []*Join             `json:"joins,omitempty" hcl:"join,block"`
    Rename      map[string][]string `json:"rename" hcl:"rename"`
    FIELDS      string              `json:"fields,omitempty" hcl:"fields"`

    TotalForce  int    `json:"total_force,omitempty" hcl:"total_force,optional"`
    MAXPAGENO   string `json:"maxpageno,omitempty" hcl:"maxpageno,optional"`
    TOTALNO     string `json:"totalno,omitempty" hcl:"totalno,optional"`
    ROWCOUNT    string `json:"rawcount,omitempty" hcl:"rawcount,optional"`
    PAGENO      string `json:"pageno,omitempty" hcl:"pageno,optional"`
    SORTBY      string `json:"sortby,omitempty" hcl:"sortby,optional"`
    SORTREVERSE string `json:"sortreverse,omitempty" hcl:"sortreverse,optional"`
}
```
It selects multiple rows with pagination. The meanings of the capital fields are:

Field | Default | Meaning in Input Data `ARGS`
--------- | ------- | -----------------------
_MAXPAGENO_ | "maxpageno" | how many pages in total
_TOTALNO_ | "totalno" | how many records in total
_ROWCOUNT_ | "rowcount" | how many record in each page
_PAGENO_ | "pageno" | return only data of the specific page
_SORTBY_ | "sortby" | sort the returned data by this
_SORTREVERSE_ | "sortreverse" | 1 to return the data in reverse

_TotalForce_ is defined in this way: 0 for not calculating total number of records; -1 for calculating; and 1 for optionally calculating. In the last case, if there is no input data for `ROWCOUNT` or `PAGENO`, there is no pagination information.

#### 2.2.6) *Delete* 

```go
type Delete struct {
    Action
}
```

It deletes a row by the primary key. 

<br />

### 2.3  *Nextpages* for Follow-up Actions

As in GraphQL, *godbi* allows an action to trigger multiple actions. *Nextpages* which is a slice of *Edge*, defines such follow-up actions:

```go
type Edge struct {
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

Parsing the JSON will build up the `Nextpages` as `map[string][]*Edge`.

<br /><br />

### 2.3  Type *Model*

`Model` contains the table and all actions on the table. 
```go
type Model struct {
	Table
	Actions map[string]interface{} `json:"actions,omitempty" hcl:"actions,optional"`
```

We can parse _Model_ from disk file `filename`:

```go
func NewModelJsonFile(filename string, custom ...map[string]Capability) (*Model, error)
```
where _custom_ would add customized actions where using action's name as the key and _Capability_ as the value. Note that

Make sure to run `Assertion` to assert the JSON structure to the right `Action` types, if to parse differently:
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
