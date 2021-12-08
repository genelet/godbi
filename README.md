# godbi

_godbi_ adds a set of high-level functions to the generic SQL handle in GO. Check *godoc* for definitions:
[![GoDoc](https://godoc.org/github.com/genelet/godbi?status.svg)](https://godoc.org/github.com/genelet/godbi)

There are three levels of usages:

- _Basic_: on raw SQL statements
- _Model_: on CRUD actions of table.
- _Graph_: on GraphQL/gRPC actions of database

The package is fully tested in MySQL and PostgreSQL.

<br /><br />

## Installation

> $ go get -u github.com/genelet/godbi
<!-- go mod init github.com/genelet/godbi -->

## Termilogy

The names of arguments passed in functions or methods are defined as follows, if not specifically explained:
Name | Type | IN/OUT | Meaning
---- | ---- | ------ | -------
*args* | `...interface{}` | IN | arguments
*ARGS* | `map[string]interface{}` | IN | input data
*extra* | `...map[string]interface{}` | IN | _WHERE_ constraints
*lists* | `[]map[string]interface{}` | OUT | data output as a slice of maps.

<br /><br />

## Chapter 1. BASIC USAGE

### 1.1  _DBI_

The `DBI` type embeds the standard SQL handle.

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

<br />

### 1.2  `DoSQL`

```go
func (*DBI) DoSQL(query string, args ...interface{}) error
```

The same as DB's `Exec`, except it returns error only.

<br />

### 1.3  `TxSQL`

```go
func (*DBI) TxSQL(query string, args ...interface{}) error
```

The same as `DoSQL` using transaction.

<br />

### 1.4   _Select_

#### 1.4.1)  `Select`

```go
func (*DBI) Select(lists *[]map[string]interface{}, query string, args ...interface{}) error
```

It runs a *Select* query and saves the result into *lists*. The data types of the rows are determined dynamically by the generic SQL handle.

<details>
    <summary>Click for example</summary>
    <p>

```go
lists := make([]map[string]interface{})
err = dbi.Select(&lists,
    `SELECT ts, id, name, len, flag, fv FROM mytable WHERE id=?`, 1234)
```

will select all rows with *id=1234* into *lists*:

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

The same as *Select* using *labels*. The interface can be either *string*
which defines renamed key name, or *[2]string* which defines renames key name and its data type like _int_, _int64_ and _string_ etc.

<details>
    <summary>Click for example</summary>
    <p>

The following example assigns key names _TS_, _id_, _Name_, _Length_, _Flag_ and _fv_, of data types _string_, _int_, _string_, _int8_, _bool_ and _float32_, to the returned rows:

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

<br />

### 1.5  _GetSQL_

If there is only one row expected, it returns the data as a map.

```go
func (*DBI) GetSQL(res map[string]interface{}, query string, labels []interface{}, args ...interface{}) error
```

<br />

### 1.6) *DBI* Example

In this example, we create table _letters_ with 3 rows, then query the data into *lists*.

<details>
    <summary>Click for Sample 1</summary>
    <p>

```go
package main

import (
    "log"
    "database/sql"
    "os"
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

    // create a new table and insert 3 rows
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

    // select all data from the table using SelectSQL
    //
    lists := make([]map[string]interface{}, 0)
    err = dbi.SelectSQL(&lists, "SELECT id, x FROM letters")
    if err != nil { panic(err) }

    log.Printf("%v", lists)

    dbi.Exec(`DROP TABLE IF EXISTS letters`)
    db.Close()
}
```

Running the example will result in

```bash
[map[id:1 x:m] map[id:2 x:n] map[id:3 x:p]]
```

</p>
</details>

<br /><br />

## Chapter 2. MODEL USAGE

The following _CRUD_ actions and _REST_ methods are defined in our package:

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
    TableName  string    `json:"table,omitempty"`  // the table name
    Pks           []string  `json:"pks,omitempty"`     // optional, the PK 
    IdAuto        string    `json:"idAuto,omitempty"` // table's auto id
    Fks           []string  `json:"fks,omitempty"`     // optional, for the FK
}
```

where _TableName_ is the table name; _Pks_ the primary key defined as a slice of columns; _IdAuto_ (optional) the column of a series number and _Fks_ (optional) the foreign key information.

_Fks_ defines a relationship between columns in two tables, which does not need to be a native foreign key:

index | meaning
----- | -------------------------
0 | the foreign table name
1 | the primary key name in the foreign table
2 | the signature name of foreign table's primary key
3 | the corresponding column of foreign table's PK in the current table
4 | the signature name of current table's primary key

Currently, to use this feature, we require table's primary key is a single column.
<br />

### 2.2  *Action*

*Action* defines an action on table, such as *CRUD*. It should implement function *RunActionContext* using the *Capability* interface:

```go
type Action struct {
    Must      []string    `json:"must,omitempty"
    Nextpages []*Nextpage `json:"nextpages,omitempty"
    Appendix  interface{} `json:"appendix,omitempty"
}
```

```go
type Capability interface {
    RunActionContext(ctx context.Context, db *sql.DB, t *Table, ARGS map[string]interface{}, extras ...map[string]interface{}) ([]map[string]interface{}, []*Nextpage, error)
}
```

where _Must_ is a slice of `NOT NULL` columns; _Nextpages_ other actions to follow after the current one is complete (for _Nextpage_, see below); and _Appendix_ optional data.

In _RunActionContext_, _ARGS_ is the input data and _extras_ a slice of constraints for the current action and all follow-up actions. This function returns the output data, the follow-up _Nextpage_s and error.

To define *extra*:

key in *extra* | meaning
-------------- | -------
has one value | an EQUAL constraint
has multiple values | an IN constraint
named *_gsql* | a raw SQL statement

The AND relation is assumed among multiple keys.

The following *CRUD* actions are pre-defined.

#### 2.2.1) *Insert*

```go
type Insert struct {
    Action
    Columns    []string      `json:"columns,omitempty" hcl:"columns,optional"`
}
```

It adds a new row into _Columns_ of table. The input data are in *ARGS* defined in _RunActionContext_.

#### 2.2.2) *Update*

```go
ype Update struct {
    Action
    Columns    []string      `json:"columns,omitempty" hcl:"columns,optional"`
    Empties    []string      `json:"empties,omitempty" hcl:"empties,optional"`
}
```

It updates row's *Columns* using the primary key. _Empties_ defines columns
which are forced to be empty or null if having no input data.

#### 2.2.3) *Insupd*

```go
type Insupd struct {
    Action
    Columns    []string      `json:"columns,omitempty" hcl:"columns,optional"`
    Uniques    []string      `json:"uniques,omitempty" hcl:"uniques,optional"`
}
```

It checks if the input data is unique using input data from columns _Uniques_. If it exists, run a *Update* otherwise *Insert*.

#### 2.2.4) *Edit*

```go
type Edit struct {
    Action
    Joins    []*Join             `json:"joins,omitempty" hcl:"join,block"`
    Columns   map[string][]string `json:"columns" hcl:"columns"`
    FIELDS   string              `json:"fields,omitempty" hcl:"fields"`
}
```

```go
type Join struct {
	Name   string `json:"name" hcl:"name,label"`
	Alias  string `json:"alias,omitempty" hcl:"alias,optional"`
	Type   string `json:"type,omitempty" hcl:"type,optional"`
	Using  string `json:"using,omitempty" hcl:"using,optional"`
	On     string `json:"on,omitempty" hcl:"on,optional"`
	Sortby string `json:"sortby,omitempty" hcl:"sortby,optional"`
}

It searches one row from a table or joint tables, using main table's primary key.

*Joins* defines the joint tables. The main table is the first element
in the list. *Join* defines the table and how it is joint. *Name* is main table's name; *Alias* the table name alias; *Type* the join type such as *JOIN*, "INNER JOIN", "LEFT JOIN" etc.; "Using" the *USING* statement in SQL; "On" the "ON* statement; and *Sortby* which appears only in the main table the column name to sort the data.
 
In *Columns*, the keys are SQL column names and keys' values are renamed labels. See 1.4.2).

*FIELDS* defines a key in the input data which, if exists, tells us
to return only a few selected columns instead of whole set of columns.
By default, _FIELDS ="fields"_.  For example, in order to return
just *user_id* and *username*, we can assign
the following value in the input data: _ARGS["fields"] = "user_id,username"_. (Note that multiple column names are concated and separated by comma.)

#### 2.2.5) *Topics*

```go
    Action
    Joins       []*Join             `json:"joins,omitempty" hcl:"join,block"`
    Columns      map[string][]string `json:"columns" hcl:"columns"`
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

It searches many rows with pagination. The capital fields are:

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

### 2.3  *Model*

*Model* contains the table and actions on the table, using the *Navigate* interface:

```go
type Model struct {
 Table
 Actions map[string]interface{} `json:"actions,omitempty" hcl:"actions,optional"`
}
```

```go
type Navigate interface {
    NonePass(action string) []string
    RunModelContext(ctx context.Context, db *sql.DB, action string, ARGS map[string]interface{}, extra ...map[string]interface{}) ([]map[string]interface{}, []*Nextpage, error)
}
```

where *Actions* defines an action name (as the key) and the associated *Capability* (as the value).

*NonePass* defines columns, whose values should not be passed to the next actions as constraints.

To parse _Model_ from the json file *filename*
containing the *CRUD* capabilities and customized capabilities *custom*:

```go
func NewModelJsonFile(filename string, custom ...map[string]Capability) (*Model, error)
```

Note that if to write own parse function, make sure to run `Assertion` to assert right `Action` types:

```go
func (self *Model) Assertion(custom ...map[string]Capability) error
```

<br />

### 2.4  *Nextpage*

As in GraphQL and gRPC, *godbi* allows action to trigger multiple actions using *Nextpage*:

```go
type Nextpage struct {
    Model      string            `json:"model"`                 // name of the next model to call  
    Action     string            `json:"action"`                // action name of the next model
    RelateItem map[string]string `json:"relateItem,omitempty"` // column name mapped to that of the next model
    Extra      map[string]string `json:"extra,omitempty"`      // manually assign these constraints
}
```

where *Models* are defined as a map between names and instance of type *Model*, and *Model* in *Nextpage* denotes the name. *Action* is the action name, *RelateItem* the map between the current data columns to next action's columns, whose values will be used as constraints, *Extra* the manually-input constraint on the next action.

Here is a use case. There are two tables, one for family and the other for children, corresponding to models named `ta` and `tb` respectively.
We search the family name in `ta`, and want to show all children as well. Technically, it means we need to run a `Topics` action on *ta*. For each row returned, we run *Topics* on *tb*, constrained by the family ID in both the tables.

So *Nextpages* of *Topics* on *ta* will look like:

<details>
    <summary>Click to show the JSON string</summary>
    <p>

```json
{
    "topics" : [
        {"model":"tb", "action":"topics", "relateItem":{"id":"id"}}
    ]
}
```

</p>
</details>

Parsing the JSON will build up a `map[string][]*Nextpage` structure.

<br />

### 2.5) Example

<details>
    <summary>Click for Example 2</summary>
    <p>

```go
package main

import (
    "log"
    "os"
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

    db.Exec(`DROP TABLE IF EXISTS testing`)
    db.Exec(`CREATE TABLE testing (id int auto_increment, x varchar(255), y varchar(255), primary key (id))`)

    table := &godbi.Table{TableName: "testing", Pks:[]string{"id"}, IdAuto:"id"}

    insert := &godbi.Insert{Columns: []string{"x","y"}}
    topics := &godbi.Topics{Columns: map[string][]string{"id":{"id","int"}, "x":{"x","string"},"y":{"y","string"}}}
    update := &godbi.Update{Columns: []string{"id","x","y"}}
    edit   := &godbi.Edit{Columns: map[string][]string{"id":{"id","int"}, "x":{"x","string"},"y":{"y","string"}}}

    args := map[string]interface{}{"x":"a","y":"b"}
    lists, _, err := insert.RunAction(db, table, args)
    if err != nil { panic(err) }
    log.Println(lists)

    args = map[string]interface{}{"x":"c","y":"d"}
    lists, _, err = insert.RunAction(db, table, args)
    if err != nil { panic(err) }
    log.Println(lists)

    args = map[string]interface{}{}
    lists, _, err = topics.RunAction(db, table, args)
    log.Println(lists)

    args = map[string]interface{}{"id":2,"x":"c","y":"z"}
    lists, _, err = update.RunAction(db, table, args)
    if err != nil { panic(err) }
    log.Println(lists)

    args = map[string]interface{}{"id":2}
    lists, _, err = edit.RunAction(db, table, args)
    log.Println(lists)

    os.Exit(0)
}
```

Running the program will result in

```bash
[map[id:1 x:a y:b]]
[map[id:2 x:c y:d]]
[map[id:1 x:a y:b] map[id:2 x:c y:d]]
[map[id:2 x:c y:z]]
[map[id:2 x:c y:z]]
```

</p>
</details>

<br /><br />

## 3. `Graph` Usage

*Graph* describes a database

```go
type Graph struct {
    *sql.DB
    Models map[string]Navigate
}
```

### 3.1 Constructor

```go
func NewGraph(db *sql.DB, s map[string]Navigate) *Graph
```

<br />

### 3.2 Run actions on models

```go
func (self *Graph) RunContext(ctx context.Context, model, action string, ARGS map[string]interface{}, extra ...map[string]interface{}) ([]map[string]interface{}, error)
```

which returns the data as *[]map[string]interface{}*, and optional error.

<br />

### 3.3) Example

<details>
    <summary>Click for Example 3</summary>
    <p>

```go
package main

import (
    "log"
    "database/sql"
    "os"
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

    ta, err := godbi.NewModelJsonFile("test_a.json")
    if err != nil { panic(err) }
    tb, err := godbi.NewModelJsonFile("test_b.json")
    if err != nil { panic(err) }

    graph := &godbi.Graph{db, map[string]godbi.Navigate{"ta":ta, "tb":tb}}

    methods := map[string]string{"LIST":"topics", "GET":"edit", "POST":"insert", "PATCH":"insupd", "PUT":"update", "DELETE":"delete"}

    var lists []map[string]interface{}
    // the 1st web requests is assumed to create id=1 to the test_a and test_b tables:
    //
    args := map[string]interface{}{"x":"a1234567","y":"b1234567","z":"temp", "child":"john"}
    if lists, err = graph.Run("ta", methods["PATCH"], args); err != nil { panic(err) }

    // the 2nd request just updates, because [x,y] is defined to the unique in ta.
    // but create a new record to tb for id=1, since insupd triggers insert in tb
    //
    args = map[string]interface{}{"x":"a1234567","y":"b1234567","z":"zzzzz", "child":"sam"}
    if lists, err = graph.Run("ta", methods["PATCH"], args); err != nil { panic(err) }

    // the 3rd request creates id=2
    //
    args = map[string]interface{}{"x":"c1234567","y":"d1234567","z":"e1234","child":"mary"}
    if lists, err = graph.Run("ta", methods["POST"], args); err != nil { panic(err) }

    // the 4th request creates id=3
    //
    args = map[string]interface{}{"x":"e1234567","y":"f1234567","z":"e1234","child":"marcus"}
    if lists, err = graph.Run("ta", methods["POST"], args); err != nil { panic(err) }

    // LIST all
    args = map[string]interface{}{}
    if lists, err = graph.Run("ta", methods["LIST"], args); err != nil { panic(err) }
    log.Printf("LIST: %v", lists)

    // GET one
    args = map[string]interface{}{"id":1}
    if lists, err = graph.Run("ta", methods["GET"], args); err != nil { panic(err) }
    log.Printf("GET: %v", lists)

    // DELETE
    extra := map[string]interface{}{"id":1}
    if lists, err = graph.Run("tb", methods["DELETE"], map[string]interface{}{"tid": 1}, extra); err != nil { panic(err) }
    if lists, err = graph.Run("tb", methods["DELETE"], map[string]interface{}{"tid": 2}, extra); err != nil { panic(err) }
    if lists, err = graph.Run("ta", methods["DELETE"], map[string]interface{}{"id":1}); err != nil { panic(err) }

    // LIST all
    args = map[string]interface{}{}
    if lists, err = graph.Run("ta", methods["LIST"], args); err != nil { panic(err) }
    log.Printf("LIST: %v", lists)

    os.Exit(0)
}
```

Running it will result in:

```bash
LIST: [map[id:1 ta_edit:[map[id:1 tb_topics:[map[child:john id:1 tid:1]] x:a1234567 y:b1234567 z:zzzzz]] x:a1234567 y:b1234567 z:zzzzz] map[id:2 ta_edit:[map[id:2 tb_topics:[map[child:mary id:2 tid:3]] x:c1234567 y:d1234567 z:e1234]] x:c1234567 y:d1234567 z:e1234] map[id:3 ta_edit:[map[id:3 tb_topics:[map[child:marcus id:3 tid:4]] x:e1234567 y:f1234567 z:e1234]] x:e1234567 y:f1234567 z:e1234]]
GET: [map[id:1 tb_topics:[map[child:john id:1 tid:1]] x:a1234567 y:b1234567 z:zzzzz]]
LIST: [map[id:2 ta_edit:[map[id:2 tb_topics:[map[child:mary id:2 tid:3]] x:c1234567 y:d1234567 z:e1234]] x:c1234567 y:d1234567 z:e1234] map[id:3 ta_edit:[map[id:3 tb_topics:[map[child:marcus id:3 tid:4]] x:e1234567 y:f1234567 z:e1234]] x:e1234567 y:f1234567 z:e1234]]
```

</p>
</details>
