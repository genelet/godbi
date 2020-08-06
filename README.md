# godbi
_godbi_ adds a set of high-level functions to the generic SQL handle in GO, for easier database executions and queries. 

Check *godoc* for definitions:
[![GoDoc](https://godoc.org/github.com/genelet/godbi?status.svg)](https://godoc.org/github.com/genelet/godbi)

There are three levels of usages:
- _Basic_: operating on raw SQL statements and stored procedures.
- _Crud_: operating on specific table and fulfilling CRUD actions.
- _Advanced_: operating on tables, called Models, as in MVC pattern in web applications, and fulfilling RESTful and GraphQL actions.


_godbi_ is an ideal replacement of ORM. It runs SQL, CRUD, RESTful and GraphQL tasks gracefully and very efficiently.
The package is fully tested in MySQL and PostgreSQL, and assumed to work with other relational databases.


<br /><br />
### Installation

> $ go get -u github.com/genelet/godbi
<!-- go mod init github.com/genelet/godbi -->


<br /><br />
## Chapter 1. BASIC USAGE


### 1.1  Type _DBI_

The `DBI` type simply embeds the standard SQL handle.
```go
package godbi

type DBI struct {
    *sql.DB          // Note this is the pointer to the handle
    LastId    int64  // read only, saves the last inserted id
    Affected  int64  // read only, saves the affected rows
}

```


#### 1.1.1) Create a new handle

```go
dbi := &DBI{DB: the_standard_sql_handle}
```

#### 1.1.2) Example

In this example, we create a MySQL handle using database credentials in the environment; then create a new table _letters_ and add 3 rows. We query the data using `SelectSQL` and put the result into `lists` as slice of maps.
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
    if err = dbi.ExecSQL(`DROP TABLE IF EXISTS letters`); err != nil { panic(err) }
    if err = dbi.ExecSQL(`CREATE TABLE letters (
        id int auto_increment primary key, x varchar(1))`); err != nil { panic(err) }
    if err = dbi.ExecSQL(`INSERT INTO letters (x) VALUES ('m')`); err != nil { panic(err) }
    if err = dbi.ExecSQL(`INSERT INTO letters (x) VALUES ('n')`); err != nil { panic(err) }
    if err = dbi.ExecSQL(`INSERT INTO letters (x) VALUES ('p')`); err != nil { panic(err) }

    // select data from the table and put them into lists
    //
    lists := make([]map[string]interface{},0)
    if err = dbi.SelectSQL(&lists, "SELECT id, x FROM letters"); err != nil { panic(err) }

    // print it
    log.Printf("%v", lists)

    dbi.ExecSQL(`DROP TABLE letters`)

    os.Exit(0)
}
```
Running this example will result in something like
```
[map[id:1 x:m] map[id:2 x:n] map[id:3 x:p]]
```


<br /><br />
### 1.2  Execution with `ExecSQL` & `DoSQL`

```go
func (*DBI) ExecSQL(query string, args ...interface{}) error
func (*DBI) DoSQL  (query string, args ...interface{}) error
```
Similar to SQL's `Exec`, these functions execute *Do*-type (e.g. _INSERT_ or _UPDATE_) queries. The difference between the two functions is that `DoSQL` runs a prepared statement and is safe for concurrent use by multiple goroutines.

For all functions in this package, the returned value is always `error` which should be checked to assert if the execution is successful.



<br /><br />
### 1.3   _SELECT_ Queries

#### 1.3.1)  `QuerySQL` & `SelectSQL`

```go
func (*DBI) QuerySQL (lists *[]map[string]interface{}, query string, args ...interface{}) error
func (*DBI) SelectSQL(lists *[]map[string]interface{}, query string, args ...interface{}) error
```
Run the *SELECT*-type query and put the result into `lists`, a slice of column name-value maps. The data types of the column are determined dynamically by the generic SQL handle. For example:
```go
lists := make([]map[string]interface{})
err = dbi.QuerySQL(&lists,
    `SELECT ts, id, name, len, flag, fv FROM mytable WHERE id=?`, 1234)
```
will select all rows with *id=1234*.
```json
    {"ts":"2019-12-15 01:01:01", "id":1234, "name":"company", "len":30, "flag":true, "fv":789.123},
    ....
```
The difference between the two functions is that `SelectSQL` runs a prepared statement.

#### 1.3.2) `QuerySQLType` & `SelectSQLType`

```go
func (*DBI) QuerySQLType (lists *[]map[string]interface{}, typeLabels []string, query string, args ...interface{}) error
func (*DBI) SelectSQLType(lists *[]map[string]interface{}, typeLabels []string, query string, args ...interface{}) error
```
They differ from the above `QuerySQL` by specifying the data types. While the generic handle could correctly figure out them in most cases, it occasionally fails because there is no exact matching between SQL typies and GOLANG types.

The following example assigns _string_, _int_, _string_, _int8_, _bool_ and _float32_ to the corresponding columns:
```go
err = dbi.QuerySQLType(&lists, []string{"string", "int", "string", "int8", "bool", "float32},
    `SELECT ts, id, name, len, flag, fv FROM mytable WHERE id=?`, 1234)
```

#### 1.3.3) `QuerySQLLabel` & `SelectSQLLable`

```go
func (*DBI) QuerySQLLabel (lists *[]map[string]interface{}, selectLabels []string, query string, args ...interface{}) error
func (*DBI) SelectSQLLabel(lists *[]map[string]interface{}, selectLabels []string, query string, args ...interface{}) error
```
They differ from the above `QuerySQL`by renaming the default column names to `electLabels` For example:
```go
lists := make([]map[string]interface{})
err = dbi.QuerySQLLabel(&lists, []string{"time stamp", "record ID", "recorder name", "length", "flag", "values"},
    `SELECT ts, id, name, len, flag, fv FROM mytable WHERE id=?`, 1234)
```
The result has the renamed keys:
```json
    {"time stamp":"2019-12-15 01:01:01", "record ID":1234, "recorder name":"company", "length":30, "flag":true, "values":789.123},
```

#### 1.3.4) `QuerySQLTypeLabel`& `SelectSQlTypeLabel`

```go
func (*DBI) QuerySQLTypeLabel (lists *[]map[string]interface{}, typeLabels []string, selectLabels []string, query string, args ...interface{}) error
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
which is similar to `SelectSQLLable` but has only single output to `res`.

#### 1.4.2) `GetArgs`

```go
func (*DBI) GetArgs(res url.Values, query string, args ...interface{}) error
```
which is similar to `SelectSQL` but has only sinlge output to `res` which uses type [url.Values](https://golang.org/pkg/net/url/). This function will be used mainly in web applications, where HTTP request data are expressed in `url.Values`.


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

Full example:
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

    if err = dbi.ExecSQL(`drop procedure if exists proc_w2`); err != nil { panic(err) }
    if err = dbi.ExecSQL(`drop table if exists letters`); err != nil { panic(err) }
    if err = dbi.ExecSQL(`create table letters(
        id int auto_increment primary key, x varchar(1))`); err != nil { panic(err) }
    if err = dbi.ExecSQL(`create procedure proc_w2(IN x0 varchar(1),OUT y0 int)
        begin
        delete from letters;
        insert into letters (x) values('m');
        insert into letters (x) values('n');
        insert into letters (x) values('p');
        insert into letters (x) values('m');
        select id, x from letters where x=x0;
        insert into letters (x) values('a');
        set y0=100;
        end`); err != nil { panic(err) }

    hash := make(map[string]interface{})
    lists := make([]map[string]interface{},0)
    if err = dbi.SelectDoProc(&lists, hash, []string{"amount"}, "proc_w2", "m"); err != nil { panic(err) }

    log.Printf("lists is: %v", lists)
    log.Printf("OUT is: %v", hash)

    dbi.ExecSQL(`drop table if exists letters`)
    dbi.ExecSQL(`drop procedure if exists proc_w2`)

    os.Exit(0)
}
```
Running the program will result in:
```
 lists is: [map[id:1 x:m] map[id:4 x:m]]
 OUT is: map[amount:100]
```


<br /><br />
## Chapter 2. CRUD USAGE

### 2.1  Type *Crud*

Type `Crud` lets us to run CRUD verbs easily on a table.  
```go
type Crud struct {
    DBI
    CurrentTable string    `json:"current_table,omitempty"`  // the current table name 
    CurrentTables []*Table `json:"current_tables,omitempty"` // optional, use multiple table JOINs in Read All 
    CurrentKey string      `json:"current_key,omitempty"`    // the single primary key of the table    
    CurrentKeys []string   `json:"current_keys,omitempty"`   // optional, if the primary key has multiple columns   
    Updated bool           // for Insupd() only, if the row is updated or new
}
```
Just to note, the 4 letters in CRUD are: 
- C: **C**reate a new row
- R: **R**ead all rows, or **R**ead one row
- U: **U**pdate a row
- D: **D**elete a row

#### 2.1.1) Create an instance

Create a new instance:
```go
crud := &godbi.Crud{DBI:dbi_created, CurrentTable:"mytable", CurrentKey:"mykey"}
```

#### 2.1.2) Example

This example creates 3 rows. Then it updates one row, reads one row and reads all rows.
```go
package main

import (
    "log"
    "net/url"
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

    // create a new instance for table "atesting"
    //
    dbi := godbi.DBI{DB:db}
    crud := &godbi.Crud{dbi, "atesting", nil, "id", nil, false}

    if err = crud.ExecSQL(`DROP TABLE IF EXISTS atesting`); err != nil { panic(err) }
    if err = crud.ExecSQL(`CREATE TABLE atesting (
        id int auto_increment, x varchar(255), y varchar(255), primary key (id))`); err != nil { panic(err) }

    // create 3 rows one by one using url.Values
    //
    hash := url.Values{}
    hash.Set("x", "a")
    hash.Set("y", "b") 
    if err = crud.InsertHash(hash); err != nil { panic(err) }
    hash.Set("x", "c")
    hash.Set("y", "d") 
    if err = crud.InsertHash(hash); err != nil { panic(err) }
    hash.Set("x", "c")
    hash.Set("y", "e")  
    if err = crud.InsertHash(hash); err != nil { panic(err) }

    // now the id is 3
    //
    id := crud.LastId
    log.Printf("last id=%d", id)
    
    // update the row of id=3, change column y to be "z"
    //
    hash1 := url.Values{}
    hash1.Set("y", "z")
    if err = crud.UpdateHash(hash1, []interface{}{id}); err != nil { panic(err) }

    // read one of the row of id=3. Only the columns x and y are reported
    //
    lists := make([]map[string]interface{}, 0)
    label := []string{"x", "y"}
    if err = crud.EditHash(&lists, label, []interface{}{id}); err != nil { panic(err) }
    log.Printf("row of id=2: %v", lists)

    // read all rows with contraint x='c'
    //
    lists = make([]map[string]interface{}, 0)
    label = []string{"id", "x", "y"}
    extra := url.Values{"x":[]string{"c"}}
    if err = crud.TopicsHash(&lists, label, extra); err != nil { panic(err) }
    log.Printf("all rows: %v", lists)

    os.Exit(0)
}
```
Running result:
```
last id=3
row of id=3: [map[x:c y:z]]
all rows: [map[id:2 x:c y:d] map[id:3 x:c y:z]]
```

<br /><br />
### 2.2  Create New Row

```go
func (*Crud) InsertHash(fieldValues url.Values) error
```
where `fieldValues` of type `url.Values` stores column's names and values. The latest inserted id will be put in `LastId` if the database driver supports it.


<br /><br />
### 2.3  Read All Rows

```go
func (*Crud) TopicsHash(lists *[]map[string]interface{}, selectPars interface{}, extra ...url.Values) error
```
where `lists` receives the query results.

#### 2.3.1) Specify which columns to be reported

Use `selectPars` which is an interface to specify which column names and types in the query. There are 4 cases:
interface | column names
--------- | ------------
 *[]string{name}* | just a list of column names
 *[][2]string{name, type}* | a list of column names and associated data types
 *map[string]string{name: label}* | rename the column names by labels
 *map[string][2]string{name: label, type}* | rename the column names to labels and use the specific types

If we don't specify type, the generic handle will decide one for us, which is most likely correct.

#### 2.3.2) Constraints

Use `extra`, which has type `url.Values` to contrain the *WHERE* statement. Currently we have supported 3 cases:
key in `extra` | meaning
--------------------------- | -------
key has only one value | an EQUAL constraint
key has multiple values | an IN constraint
key is named *_gsql* | a raw SQL statement
among multiple keys | AND conditions.

#### 2.3.3) Use multiple JOIN tables

The _R_ verb will use a JOIN SQL statement from related tables, if `CurrentTables` of type `Table` exists in `Crud`. Type `Table` is usually parsed from JSON.
```go
type Table struct { 
    Name string   `json:"name"`             // name of the table
    Alias string  `json:"alias,omitempty"`  // optional alias of the table
    Type string   `json:"type,omitempty"`   // INNER or LEFT, how the table is joined
    Using string  `json:"using,omitempty"`  // optional, joining by USING table name
    On string     `json:"on,omitempty"`     // optional, joining by ON condition
    Sortby string `json:"sortby,omitempty"` // optional column to sort, only applied to the first table
}
```
The tables in `CurrentTables` should be arranged with correct orders. Use the following function to create a SQL logic: 
```go
func TableString(tables []*Table) string
```
For example:
```go
str := `[
    {"name":"user_project", "alias":"j"},
    {"name":"user_component", "alias":"c", "type":"INNER", "using":"projectid"},
    {"name":"user_table", "alias":"t", "type":"LEFT", "on":"c.tableid=t.tableid"}]`
tables := make([]*Table, 0)
if err := json.Unmarshal([]byte(str), &tables); err != nil { panic(err) }
log.Printf("%s", TableString())
```
will output
```sql
user_project j
INNER JOIN user_component c USING (projectid)
LEFT JOIN user_table t ON (c.tableid=t.tableid)
```

By combining `selectPars` and `extra`, we can construct sophisticate search queries. More use cases will be discussed in _Advanced Usage_ below.


<br /><br />
### 2.4  Read One Row

```go
func (*Crud) EditHash(lists *[]map[string]interface{}, editPars interface{}, ids []interface{}, extra ...url.Values) error
```
This will select rows having the specific primary key (*PK*) values `ids` and being constrained by `extra`. The query result is output to `lists` with columns defined in `editPars`. For the slice of interface `ids`:
- if PK is a single column, `ids` should be a slice of targeted PK values
  - to select a single PK equaling to 1234, just use `ids = []int{1234}`
- if PK has multiple columns, i.e. `CurrentKeys` exists, `ids` should be a slice of value arrays.


<br /><br />
### 2.5  Update Row

```go
func (*Crud) UpdateHash(fieldValues url.Values, ids []interface{}, extra ...url.Values) error
```
The rows having `ids` as PK and `extra` as constraint will be updated. The columns and the new values are defined in `fieldValues`.


<br /><br />
### 2.6  Create or Update Row

```go
func (*Crud) InsupdTable(fieldValues url.Values, uniques []string) error
```
This function is not a part of CRUD, but is implemented as *PATCH* method in *http*. When we try create a row, it may already exist. If so, we will update it instead. The uniqueness is determined by `uniques` column names. The field `Updated` will tell if the verb is updated or not. 


<br /><br />
### 2.7  Delete Row

```go
func (*Crud) DeleteHash(extra ...url.Values) error
```
This function deletes rows using constrained `extra`.



<br /><br />
## Chapter 3. ADVANCED USAGE

godbi allows us to construct database *model* as in the MVC Pattern in web applications, and to build RESTful API easily. The RESTful web actions are associated with the database CRUD verbs as the following:

HTTP METHOD | RESTful URL | CRUD | Function in godbi 
----------- | ----------- | ---- | -----------------
GET         | webHandler | R All | Topics
GET         | webHandler/ID | R One | Edit
POST        | webHandler | C | Insert
PUT         | webHandler | U | Update
PATCH       | webHandler | NA | Insupd
DELETE      | webHandler | D | Delete

Futhermore, we have set up type `Schema` for whole database schema, which allow us to build multiple API endpoints at once. 


<br /><br />
### 3.1  Type *Model*

```go
type Model struct {
    Crud
    Navigate                          // interface has methods to implement 

    Nextpages      map[string][]*Page `json:"nextpages,omitempty"`       // to call other models' verbs
    CurrentIdAuto  string             `json:"current_id_auto,omitempty"` // this table has an auto id
    InsertPars     []string           `json:"insert_pars,omitempty"`     // columns to insert in C
    UpdatePars     []string           `json:"update_pars,omitempty"`     // columns to update in U
    InsupdPars     []string           `json:"insupd_pars,omitempty"`     // unique columns in PATCH
    EditPars       []string           `json:"edit_pars,omitempty"`       // columns to query in R (one)
    TopicsPars     []string           `json:"topics_pars,omitempty"`     // columns to query in R (all)
    TopicsHashPars map[string]string  `json:"topics_hash,omitempty"`     // columns to rename in R (all)
    TotalForce     int                `json:"total_force,omitempty"`     // if to calculate total coutns in R (all)
    
    // The following fields are just variable names to pass in a web request,
    // default to themselves. e.g. "empties" for "Empties", "maxpageno" for Maxpageno etc.
    Empties        string             `json:"empties,omitempty"`         // columns are updated to NULL if no input 
    Fields         string             `json:"fields,omitempty"`          // use this smaller set of columns in R 

    // these fields are for pagination.
    Maxpageno      string             `json:"maxpageno,omitempty"`       // total page no. 
    Totalno        string             `json:"totalno,omitempty"`         // total item no. 
    Rowcount       string             `json:"rawcount,omitempty"`        // counts per page  
    Pageno         string             `json:"pageno,omitempty"`          // current page no. 
    Sortreverse    string             `json:"sortreverse,omitempty"`     // if reverse sorting
    Sortby         string             `json:"sortby,omitempty"`          // sorting column
}
```
where the interface `Navigate`:
```go
type Navigate interface {
    SetArgs(url.Values)                             // set http request data
    GetArgs(...bool)     url.Values                 // get http request data for the nextpage to use

    SetDB(*sql.DB)                                  // set the database handle 

    SetActions(map[string]func(...url.Values)error) // set new map between name and action function
    RunAction(string, ...url.Values) error          // run an action by name

    GetLists()           []map[string]interface{}   // get result after an action
    GetNextpages(string) []*Page                    // get the nextpages
}
```
In *godbi*, the `Model` type has already implemented the 4 methods. 


#### 3.1.1) Constructor `NewModel`

A `Model` instance can be parsed from JSON file on disk:
```go
func NewModel(filename string) (*Model, error)
```
where `filename` is the file name. Please check the tags on struct field declarations in `Crud` and `Model`:

Field in Model | JSON variable | Database Table
-------------- | ------------- | --------------
CurrentTable | current_table | the current table name 
CurrentTables | current_tables | optional, use multiple table JOINs in Read All
CurrentKey | current_key | the single primary key of the table    
CurrentKeys | current_keys | optional, if the primary key has multiple columns  
CurrentIdAuto  | current_id_auto | this table has an auto id
InsertPars     | insert_pars | columns to insert in C
UpdatePars     | update_pars | columns to update in U
InsupdPars     | insupd_pars | unique columns in PATCH
EditPars       | edit_pars | columns to query in R (one)
TopicsPars     | topics_pars | columns to query in R (all)
TopicsHashPars | topics_hash | columns to rename in R (all)
TotalForce     | total_force | if to calculate total counts in R (all)

#### 3.1.2) Feed Database Handle and Input Data

Use 
```go
func (*Model) SetDB(db *sql.DB)
func (*Model) SetArgs(args url.Values)
```
to set database handle, `db`, and input data, `args` like http request's *Form* into `Model`:

#### 3.1.3) Returning Data

After we have run an action on the model, we can retrieve data using 
```go
(*Model) GetLists()
```
The input data `args` may has been modified too. To get it back use
```go
(*Model) GetArgs(ignore ...bool)
```
`ignore=true` will turn off the pagination information of *totalno*, *maxpageno* and *rowcount*.

#### 3.1.4) Http METHOD: GET (read all)

```
func (*Model) Topics(extra ...url.Values) error
```
It selects all records in model's table, constrained optionally by `extra`. 

If variable `rowcount` (*number of records per page*) is set in `args`, and field `TotalForce` is not 0, then pagination will be triggered. The total count and total pages will be calculatd and put back in `args`. `TotalForce` defines how to calculate the total count.
Value | Meaning
----- | -------
<-1  | use ABS(TotalForce) as the total count 
-1   | always calculate the total count
0    | don't calculate the total count
&gt; 0  | calculate only if the total count is not passed in `args`

#### 3.1.4) Http METHOD: GET (read one)

> Note: the database table columns in the following RESTful actions are already defined in the model.

```
func (*Model) Edit(extra ...url.Values) error
```
It selects one records according to the PK value in the input data, constrained optionally by `extra`. 

#### 3.1.5) Http METHOD: POST

```
func (*Model) Insert(extra ...url.Values) error
```
It inserts a new record using the input data. If `extra` is passed in, it will override the input data.

#### 3.1.6) Http METHOD: PUT

```
func (*Model) Update(extra ...url.Values) error
```
It updates a row according to the PK and the input data, constrained optionally by `extra`.

#### 3.1.7) Http METHOD: PATCH

```
func (*Model) Insupd(extra ...url.Values) error
```
It inserts or updates a row using the input data, constrained optionally by `extra`.

#### 3.1.8) Http METHOD: DELETE

```
func (*Model) Delete(extra ...url.Values) error
```
It rows constrained by `extra`. For this function, the input data will NOT be used.

#### 3.1.9）Example

This example shows how to run RESTful actions on a model
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
	model.CurrentIdAuto = "id"
	model.InsertPars    = []string{     "x","y"}
	model.TopicsPars    = []string{"id","x","y"}
	model.UpdatePars    = []string{"id","x","y"}
	model.EditPars      = []string{"id","x","y"}

	args["x"] = []string{"a"}
	args["y"] = []string{"b"}
	if err := model.Insert(); err != nil { panic(err) }
	log.Println(model.LastId)

	args["x"] = []string{"c"}
	args["y"] = []string{"d"}
	if err := model.Insert(); err != nil { panic(err) }
	log.Println(model.LastId)

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
```
1
2
[map[id:1 x:a y:b] map[id:2 x:c y:d]]
[map[id:2 x:c y:z]]
```

<br /><br />
### 3.2 Action by Name

Besides running action directly by calling its method, we can run it alternatively by calling its string name. This is important in web application where 
the server is open to many different user actions and need to call particular one dynamically according to query string.

To achieve this, we need to set all actions for the model using
```go
// SetActions: set new map between name and action function
SetActions(map[string]func(...url.Values)error)
```
which is a map between the name and action function (a closure).

For example, to map string "topics" to method `Topics` of instance `model`,
we can do
```go
actions["name"] = func(extra ...) error { return model.Topics(extra...) }
```
here `Topics` behaves as a closure.

To actually run an action using its name, use
```go
RunAction(string, ...url.Values) error
```
For example, the instance is named `model` and the action map is `actions`:
```go
// defining the action map 
actions := make(map[string]func(...url.Values)error)
actions["topics"] = func(extra ...url.Values) { return  model.Topics(extra...) }
model.SetActions(actions)
err := model.RunAction("topics", extra...)
```

<br /><br />
### 3.3 Definition of *Next Pages*

As in GraphQL and gRCP, action on a model could trigger multiple actions on other models. godbi supports this feature, called *Nextpage*.

Here is the type:
```go
type Page struct {
    Model      string            `json:"model"`                 // name of the next model to call  
    Action     string            `json:"action"`                // action name of the next model
    RelateItem map[string]string `json:"relate_item,omitempty"` // column name mapped to that of the next model
    Manual     map[string]string `json:"manual,omitempty"`      // manually assign these constraints
}
```

The behavior of next pages is usually parsed from JSON. Here is example. Assuming there are two tables, one for family and the other for children,
corresponding to two models `ta` and `tb` respectively.

Everytime when we run a RESTful action on `ta`, we always trigger a similiar action on `tb`: when we list the family name, we want to show all children under the familiy name as well. Technically, it means that running `Topics` on `ta` will trigger `Topics` on `tb`, constrained by the association of family's ID in both the tables. The same is true for 
`Edit` and `Insert`. So for the family model, its `Nextpages` will look like
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
Parsing it will result in the `map[string][]*Page` data structure. In godbi, we build up such kind of relationship once in JSON and let the package to run for us.

In case of any change in the business logic, we can modify the JSON file, which is much cleaner and easier to do than other tools like ORM.

Now let's go to `Schema` for the usagage of `Nextpages`.

<br /><br />
### 3.4 Type `Schema`

Because models are allowed to interact with each other, we introduc type `Schema` which handles the whole database schema at once:
```go
type Schema struct {
    Models  map[string]Navigate
}
```
where keys in the map are all model names.

`Schema` implement the `Run` method which is ideal for RESTful requests.
```go
func (self *Schema) Run(model, action string, args url.Values, db *sql.DB, extra ...url.Values) ([]map[string]interface{}, error)
```
We pass in the string names of model and action, the input data, the database handle, and optional extra parameters, this function runs the action and returns the results.

Here is a full example that covers all knowledges in Chapter 3.
```go
```
