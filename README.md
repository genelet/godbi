# godbi

_godbi_ adds a set of high-level functions to the generic SQL handle in GO, for easier database executions and queries. 

Check *godoc* for definitions:
[![GoDoc](https://godoc.org/github.com/genelet/godbi?status.svg)](https://godoc.org/github.com/genelet/godbi)

There are three levels of usages:

- _Basic_: operating on raw SQL statements and stored procedures.
- _Model_: operating on specific table and fulfilling CRUD actions, as *Model* in MVC pattern.
- _Schema_: operating on whole database schema and fulfilling RESTful and GraphQL actions.

_godbi_ is an ideal replacement of ORM. It runs SQL, CRUD, RESTful and GraphQL tasks gracefully and very efficiently.
The package is fully tested in MySQL and PostgreSQL, and assumed to work with other relational databases.

<br /><br />

### Installation

> $ go get -u github.com/genelet/godbi
<!-- go mod init github.com/genelet/godbi -->

### Termilogy

The names of arguments passed in functions or methods in this package are defined as follows, if not specifically explained:
Name | Type | IN/OUT | Where | Meaning
---- | ---- | ------ | ----- | -------
*args* | `...interface{}` | IN | `DBI` | single-valued interface slice, possibly empty
*args* | `url.Values` | IN | `Model` | via SetArgs() to set input data
*args* | `url.Values` | IN | `Schema` | input data passing to Run()
*extra* | `url.Values` | IN | `Model`,`Schema` | WHERE constraints; single value - EQUAL,  multi value - IN
*lists* | `[]map[string]interface{}` | OUT | all | output as slice of rows; each row is a map.
*res* | `map[string]interface{}` | OUT | `DBI` | output for one row

<br /><br />

## Chapter 1. BASIC USAGE

### 1.1  Type _DBI_

The `DBI` type simply embeds the standard SQL handle.

```go
package godbi

type DBI struct {
    *sql.DB          // Note this is the pointer to the handle
    LastID    int64  // read only, saves the last inserted id
    Affected  int64  // read only, saves the affected rows
}

```

#### 1.1.1) Create a new handle

```go
dbi := &DBI{DB: the_standard_sql_handle}
```

#### 1.1.2) Example

In this example, we create a MySQL handle using database credentials in the environment; then create a new table _letters_ and add 3 rows. We query the data using `SelectSQL` and put the result into `lists` as slice of maps.

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
    _, err = dbi.Exec(`DROP TABLE IF EXISTS letters`)
    if err != nil { panic(err) }
    _, err = dbi.Exec(`CREATE TABLE letters (
        id int auto_increment primary key,
        x varchar(1))`)
    if err != nil { panic(err) }
    _, err = dbi.Exec(`insert into letters (x) values ('m')`)
    if err != nil { panic(err) }
    _, err = dbi.Exec(`insert into letters (x) values ('n')`)
    if err != nil { panic(err) }
    _, err = dbi.Exec(`insert into letters (x) values ('p')`)
    if err != nil { panic(err) }

    // select all data from the table
    lists := make([]map[string]interface{},0)
    sql := "SELECT id, x FROM letters"
    err = dbi.SelectSQL(&lists, sql)
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

Similar to SQL's `Exec`, `DoSQL` executes *Do*-type (e.g. _INSERT_ or _UPDATE_) queries. It runs a prepared statement and may be safe for concurrent use by multiple goroutines.

For all functions in this package, the returned value is always `error` which should be checked to assert if the execution is successful.

<br /><br />

### 1.3   _SELECT_ Queries

#### 1.3.1)  `SelectSQL`

```go
func (*DBI) SelectSQL(lists *[]map[string]interface{}, query string, args ...interface{}) error
```

Run the *SELECT*-type query and put the result into `lists`, a slice of column name-value maps. The data types of the column are determined dynamically by the generic SQL handle.

<details>
    <summary>Click for example</summary>
    <p>

```go
lists := make([]map[string]interface{})
err = dbi.DoSQL(&lists,
    `SELECT ts, id, name, len, flag, fv FROM mytable WHERE id=?`, 1234)
```

will select all rows with *id=1234*.

```json
    {"ts":"2019-12-15 01:01:01", "id":1234, "name":"company", "len":30, "flag":true, "fv":789.123},
    ....
```

</p>
</details>

`SelectSQL` runs a prepared statement.

#### 1.3.2) `SelectSQLType`

```go
func (*DBI) SelectSQLType(lists *[]map[string]interface{}, typeLabels []string, query string, args ...interface{}) error
```

They differ from the above `SelectSQL` by specifying the data types. While the generic handle could correctly figure out them in most cases, it occasionally fails because there is no exact matching between SQL types and GOLANG types.

The following example assigns _string_, _int_, _string_, _int8_, _bool_ and _float32_ to the corresponding columns:

```go
err = dbi.SelectSQLType(&lists, []string{"string", "int", "string", "int8", "bool", "float32},
    `SELECT ts, id, name, len, flag, fv FROM mytable WHERE id=?`, 1234)
```

#### 1.3.3) `SelectSQLLabel`

```go
func (*DBI) SelectSQLLabel(lists *[]map[string]interface{}, selectLabels []string, query string, args ...interface{}) error
```

They differ from the above `SelectSQL` by renaming the default column names to `selectLabels`.

<details>
    <summary>Click for example</summary>
    <p>

```go
lists := make([]map[string]interface{})
err = dbi.querySQLLabel(&lists, []string{"time stamp", "record ID", "recorder name", "length", "flag", "values"},
    `SELECT ts, id, name, len, flag, fv FROM mytable WHERE id=?`, 1234)
```

The result has the renamed keys:

```json
    {"time stamp":"2019-12-15 01:01:01", "record ID":1234, "recorder name":"company", "length":30, "flag":true, "values":789.123},
```

</p>
</details>

#### 1.3.4) `SelectSQlTypeLabel`

```go
func (*DBI) SelectSQLTypeLabel(lists *[]map[string]interface{}, typeLabels []string, selectLabels []string, query string, args ...interface{}) error
```

These functions re-assign both data types and column names in the queries.

<br /><br />

### 1.4  Query Single Row

In some cases we know there is only one row from a query.

#### 1.4.1) `GetSQLLable`

```go
func (*DBI) GetSQLLabel(res map[string]interface{}, query string, selectLabels []string, args ...interface{}) error
```

which is similar to `SelectSQLLabel` but has only single output to `res`.

#### 1.4.2) `GetArgs`

```go
func (*DBI) GetArgs(res url.Values, query string, args ...interface{}) error
```

which is similar to `SelectSQL` but has only single output to `res` which uses type [url.Values](https://golang.org/pkg/net/url/). This function will be used mainly in web applications, where HTTP request data are expressed in `url.Values`.

<br /><br />

### 1.5  Stored Procedure

_godbi_ runs stored procedures easily as well.

#### 1.5.1) `DoProc`

```go
func (*DBI) DoProc(res map[string]interface{}, names []string, proc_name string, args ...interface{}) error
```

It runs a stored procedure `proc_name` with IN data in `args`. The OUT data will be placed in `res` using `names` as its keys. Note that the OUT variables should have been defined separately in `proc_name`.

If the procedure has no OUT to receive, just assign `names` to be `nil`.

#### 1.5.2) `SelectDoProc`

```go
func (*DBI) SelectDoProc(lists *[]map[string]interface{}, res map[string]interface{}, names []string, proc_name string, args ...interface{}) error
```

Similar to `DoProc` but it receives _SELECT_-type query data into `lists`, providing `proc_name` contains such a query.

<details>
    <summary>Click for full example</summary>
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

    dbi.Exec(`drop procedure if exists proc_w2`)
    dbi.Exec(`drop table if exists letters`)
    _, err = dbi.Exec(`create table letters(id int auto_increment primary key, x varchar(1))`)
    if err != nil { panic(err) }

    _, err = dbi.Exec(`create procedure proc_w2(IN x0 varchar(1),OUT y0 int)
        begin
        delete from letters;
        insert into letters (x) values('m');
        insert into letters (x) values('n');
        insert into letters (x) values('p');
        insert into letters (x) values('m');
        select id, x from letters where x=x0;
        insert into letters (x) values('a');
        set y0=100;
        end`)
    if err != nil { panic(err) }

    sql := `proc_w2`
    hash := make(map[string]interface{})
    lists := make([]map[string]interface{},0)
    err = dbi.SelectDoProc(&lists, hash, []string{"amount"}, sql, "m")
    if err != nil { panic(err) }

    log.Printf("lists is: %v", lists)
    log.Printf("OUT is: %v", hash)

    dbi.Exec(`drop table if exists letters`)

    os.Exit(0)
}
```

Running the program will result in:

```bash
 lists is: [map[id:1 x:m] map[id:4 x:m]]
 OUT is: map[amount:100]
```

</p>
</details>

<br /><br />

## Chapter 2. MODEL USAGE

*godbi* allows us to construct *model* as in the MVC Pattern in web applications, and to build RESTful API easily. The CRUD verbs on table are defined to be:
C | R | U | D
---- | ---- | ---- | ----
create a new row | read all rows, or read one row | update a row | delete a row

The RESTful web actions are associated with the CRUD verbs as follows:

<details>
    <summary>Click for RESTful vs CRUD</summary>
    <p>

HTTP METHOD | Web URL | CRUD | Function in godbi
----------- | ------- | ---- | -----------------
GET         | webHandler | R All | Topics
GET         | webHandler/ID | R One | Edit
POST        | webHandler | C | Insert
PUT         | webHandler | U | Update
PATCH       | webHandler | NA | Insupd
DELETE      | webHandler | D | Delete

</p>
</details>

<br /><br />

### 2.1  Type *Table*

*godbi* uses JSON to express the CRUD fields and logic. There should be one, and
only one, JSON (file or string) assigned to each database table. The JSON is designed only once. In case of any change in the business logic, we can modify it, which is much cleaner and easier to do than changing program code, as in ORM.

Here is the `Table` type:

```go
    CurrentTable   string             `json:"current_table,omitempty"`   // the current table name
    CurrentTables  []*Join            `json:"current_tables,omitempty"`  // optional, use multi table JOINs in Read All
    CurrentKey     string             `json:"current_key,omitempty"`     // the single primary key of the table
    CurrentKeys    []string           `json:"current_keys,omitempty"`    // optional, if the PK has multiple columns
    CurrentIDAuto  string             `json:"current_id_auto,omitempty"` // this table has an auto id
    InsertPars     []string           `json:"insert_pars,omitempty"`     // columns to insert in C
    UpdatePars     []string           `json:"update_pars,omitempty"`     // columns to update in U
    InsupdPars     []string           `json:"insupd_pars,omitempty"`     // unique columns in PATCH
    EditPars       []interface{}      `json:"edit_pars,omitempty"`       // columns to query in R (one)
    EditHash   map[string]interface{} `json:"edit_hash,omitempty"`       // R(a) with specific types and labels
    TopicsPars     []interface{}      `json:"topics_pars,omitempty"`     // columns to query in R (all)
    TopicsHash map[string]interface{} `json:"topics_hash,omitempty"`     // R(a) with specific types and labels
    TotalForce     int                `json:"total_force,omitempty"`     // if to calculate total counts in R(a)

    Nextpages      map[string][]*Page `json:"nextpages,omitempty"`       // to call other models' verbs

    // The following fields are just variable names to pass in a web request,
    // default to themselves. e.g. "empties" for "Empties", "maxpageno" for Maxpageno etc.
    Empties        string             `json:"empties,omitempty"`         // columns are updated to NULL if no input
    Fields         string             `json:"fields,omitempty"`          // use this smaller set of columns in R
    // the following fields are for pagination.
    Maxpageno      string             `json:"maxpageno,omitempty"`       // total page no.
    Totalno        string             `json:"totalno,omitempty"`         // total item no.
    Rowcount       string             `json:"rowcount,omitempty"`        // counts per page
    Pageno         string             `json:"pageno,omitempty"`          // current page no.
    Sortreverse    string             `json:"sortreverse,omitempty"`     // if reverse sorting
    Sortby         string             `json:"sortby,omitempty"`          // sorting column
}
```

And here is explanation of the fields:

<details>
    <summary>Click to Show Fields in Model</summary>
    <p>

Field in Model | JSON variable | Database Table
-------------- | ------------- | --------------
CurrentTable | current_table | the current table name
CurrentTables | current_tables | optional, use multiple table JOINs in Read All
CurrentKey | current_key | the single primary key of the table
CurrentKeys | current_keys | optional, if the primary key has multiple columns  
CurrentIDAuto  | current_id_auto | this table has an auto id
InsertPars     | insert_pars | columns to insert in C
UpdatePars     | update_pars | columns to update in U
InsupdPars     | insupd_pars | unique columns in PATCH
EditPars       | edit_pars | columns to query in R (one)
TopicsPars     | topics_pars | columns to query in R (all)
TotalForce     | total_force | if to calculate total counts in R (all)

</p>
</details>

#### 2.1.1) *Read* with specific types and/or names

While in most cases we *Read* by simple column slice, i.e. *EditPars* & *TopicsPars*,
occasionally we need specific names and types in output. Here is what *godbi* will do
in case of existence of *EditHash* or/and *TopicsHash*.

<details>
    <summary>Click to show <em>EditPars</em>, <em>EditHash</em>, <em>TopicsPars</em> and <em>TopicsHash</em></summary>
    <p>

interface | variable | column names
--------- | -------- | ------------
 *[]string{name}* | EditPars, TopicsPars | just a list of column names
 *[][2]string{name, type}* | EditPars, TopicsPars | column names and their data types
 *map[string]string{name: label}* | EditHash, TopicsHash | rename the column names by labels
 *map[string][2]string{name: label, type}* | EditHash, TopicsHash | rename and use the specific types

</p>
</details>

#### 2.1.2) Read all with multiple joined tables

If `CurrentTables` of type `[]*Join` exists in `Table`. The _R_ verb will use a SQL statement
from multi relational tables.

```go
type Join struct { 
    Name string   `json:"name"`             // name of the table
    Alias string  `json:"alias,omitempty"`  // optional alias of the table
    Type string   `json:"type,omitempty"`   // INNER or LEFT, how the table is joined
    Using string  `json:"using,omitempty"`  // optional, joining by USING table name
    On string     `json:"on,omitempty"`     // optional, joining by ON condition
    Sortby string `json:"sortby,omitempty"` // optional column to sort, only applied to the first table
}
```

The first element in *CurrentTables* is the current table and the rest for INNER JOIN or LEFT JOIN tables.

<details>
    <summary>Click for Table example</summary>
    <p>

```json
[
    {"name":"user_project", "alias":"j"},
    {"name":"user_component", "alias":"c", "type":"INNER", "using":"projectid"},
    {"name":"user_table", "alias":"t", "type":"LEFT", "on":"c.tableid=t.tableid"}
]
```
is equivalent to

```sql
user_project j
INNER JOIN user_component c USING (projectid)
LEFT JOIN user_table t ON (c.tableid=t.tableid)
```

</p>
</details>

#### 2.1.3) Pagination

We have define a few variable names whose values can be passed in input data, to make *Read All* in pagination.

First, use `TotalForce` to define how to calculate the total row count.

<details>
    <summary>Click for meaning of *TotalForce*</summary>
    <p>

Value | Meaning
----- | -------
<-1  | use ABS(TotalForce) as the total count
-1   | always calculate the total count
0    | don't calculate the total count
&gt; 0  | calculate only if the total count is not passed in `args`

</p>
</details>

If variable `rowcount` (*number of records per page*) is set in input, and field `TotalForce` is not 0, then pagination will be triggered. The total count and total pages will be calculated and put back in variable names `totalno` and `maxpageno`. For consecutive requests, we should attach values of *pageno*, *totalno* and *rowcount* to get the
right page back.

By combining *TopicsHash*, *CurrentTables* and the pagination variables, we can build up quite sophisticated SQLs for most queries.

#### 2.1.4) Definition of *Next Pages*

As in GraphQL and gRCP, *godbi* allows an action to trigger multiple actions on other models. To what actions
on other models will get triggered, define *Nextpages* in *Table*.

Here is type *Page*:

```go
type Page struct {
    Model      string            `json:"model"`                 // name of the next model to call  
    Action     string            `json:"action"`                // action name of the next model
    RelateItem map[string]string `json:"relate_item,omitempty"` // column name mapped to that of the next model
    Manual     map[string]string `json:"manual,omitempty"`      // manually assign these constraints
}
```

Assume there are two tables, one for family and the other for children, corresponding to two models `ta` and `tb` respectively.

When we *GET* the family name, we'd like to show all children under the family name as well. Technically, it means that running `Topics` on `ta` will trigger `Topics` on `tb`, constrained by the association of family's ID in both the tables. The same is true for `Edit` and `Insert`. So for the family model, its `Nextpages` will look like

<details>
    <summary>Click to show the JSON string</summary>
    <p>

```json
{
    "insert" : [
        {"model":"tb", "action":"insert", "relate_item":{"id":"id"}}
    ],
    "insupd" : [
        {"model":"tb", "action":"insert", "relate_item":{"id":"id"}}
    ],
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

Parsing it will result in `map[string][]*Page`. *godbi* will run all the next pages automatically in chain.

<br /><br />

### 2.2  Type *Model*

```go
type Model struct {
    DBI
    Table
    Navigate                                        // interface has methods to implement
    Actions   map[string]func(...url.Values) error  // action name to closure map
    Updated
```

where `Actions` is an action name to action closure map; the interface `Navigate` is:

```go
type Navigate interface {
    SetArgs(url.Values)                            // set http request data
    SetDB(*sql.DB)                                 // set the database handle
    GetAction(string)   func(...url.Values) error  // get function by action name
    GetLists()          []map[string]interface{}   // get result after an action
}
```

In *godbi*, the `Model` type has already implemented the 4 methods.

#### 2.2.1) Constructor `NewModel`

A `Model` instance can be parsed from JSON file on disk:

```go
func NewModel(filename string) (*Model, error)
```

where `filename` is the file name.

#### 2.2.2) Set Database Handle and Input Data

Use

```go
func (*Model) SetDB(db *sql.DB)
func (*Model) SetArgs(args url.Values)
```

to set database handle `db`, and input data `args`. The input data is of type *url.Values*.
In web applications, this is *Form* from http request in `net/http`.


#### 2.2.3) Optional Constraints

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
