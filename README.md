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
func (*Crud) DeleteHash(ids []interface{}, extra ...url.Values) error
```
This function deletes rows specified by `ids` and constrained by `extra`.



<br /><br />
## Chapter 3. ADVANCED USAGE

`Model` constructs a database *model*, as in the MVC Pattern, from JSON and runs CRUD actions.

<br /><br />
### 3.1  Type *Model*

```go
type Model struct {
    Crud
    Navigate                         // interface has methods to implement 
    ARGS   url.Values                // store http request data 
    LISTS  []map[string]interface{}  // store output 

    // all the following fields will be parsed from JSON
    Nextpages map[string][]*Page     `json:"nextpages,omitempty"`       // to call other models' verbs
    CurrentIdAuto  string            `json:"current_id_auto,omitempty"` // this table has an auto id
    InsertPars     []string          `json:"insert_pars,omitempty"`     // columns to insert in C
    UpdatePars     []string          `json:"update_pars,omitempty"`     // columns to update in U
    InsupdPars     []string          `json:"insupd_pars,omitempty"`     // unique columns in PATCH
    EditPars       []string          `json:"edit_pars,omitempty"`       // columns to query in R (one)
    TopicsPars     []string          `json:"topics_pars,omitempty"`     // columns to query in R (all)
    TopicsHashPars map[string]string `json:"topics_hash,omitempty"`     // columns to rename in R (all)
    TotalForce     int               `json:"total_force,omitempty"`     // if to calculate total coutns in R (all)
    
    // The following fields are just variable names to pass in a web request,
    // default to themselves. e.g. "empties" for "Empties", "maxpageno" for Maxpageno etc.
    Empties        string            `json:"empties,omitempty"`         // columns are updated to NULL if no input 
    Fields         string            `json:"fields,omitempty"`          // use this smaller set of columns in R 
    // these fields are for pagination. Their values have meanings
    Maxpageno      string            `json:"maxpageno,omitempty"`       // total page no. 
    Totalno        string            `json:"totalno,omitempty"`         // total item no. 
    Rowcount       string            `json:"rawcount,omitempty"`        // counts per page  
    Pageno         string            `json:"pageno,omitempty"`          // current page no. 
    Sortreverse    string            `json:"sortreverse,omitempty"`     // if reverse sorting
    Sortby         string            `json:"sortby,omitempty"`          // sorting column
}
```
where the interface `Navigate`:
```go
type Navigate interface {
    GetLists() []map[string]interface{} // get lists
    GetArgs() url.Values                // get http request data for the nextpage to use
    GetNextpages(string) []*Page        // get the nextpages
    UpdateModel(*sql.DB, url.Values)    // set up the database handle and the http request data
}
```
In *godbi*, the `Model` type has already implemented the 4 methods. 


Type `Page`:
```go
type Page struct {
    Model      string            `json:"model"`                // name of the next model to call  
    Action     string            `json:"action"`               // action name of the next model
    RelateItem map[string]string `json:"relate_item,omitempty" // column name mapped to that of the next model
    Manual     map[string]string `json:"manual,omitempty"`     // manually assign these constraints
}
```



####  3.1.1) What is `Model`?

`Model` generatea RESTful & GraphQL APIs in web application, for the whole database schema at once! 
It has implemeted the standard CRUD verbs, and can include custom web actions too. So is ready to generate RESTful APIs but also ready for use in
any web development envirionmen.

`Model` uses JSON to initiate, which has great advantage over ORM where one has to program sophisticate and cubsome logic
manually in code.

To create and use models of a schema, follow the 5 steps:

- Construct one JSON file for each table, and create an initial `Model` instance
- Build `Schema` from all the models in the database, and assign them to each model. 
- During the run time, e.g. when serving a http request, use `UpdateModel(db, ARGS)` to dynamically set up database handle and to feed client's *Form* data
- Run an RESTful action. If the model has `Nextpages` defined for this action, the system will run other model's action automatically. 
- Use `GetLists()` to get the result.

### 3.1.2) Create `Model`

Use
```go
func NewModel(filename string) (*Model, error)
```
to create a new `Model` instance, where `filename` is the JSON file for this table. 

Here is an example, assuming we have 2 MySQL tables in the schema,  *test_a* & "test_b*:
```sql
CREATE TABLE test_a (
  id int auto_increment not null primary key, 
  x varchar(8),
  y varchar(8),
  z varchar(8)
);

CREATE TABLE test_b (
  tid int auto_increment not null primary key,
  child varchar(8),
  id int
);
```
The associated JSON file for *test_a.json*:
```json
{
    "nextpages" : {
        "topics" : [
            {"model":"tb", "action":"edit", "relate_item":{"id":"id"}}
        ]
    }

    "current_table" : "test_a",
    "current_key" : "id",
    "current_id_auto" : "id",
    "insupd_pars" : ["x","y"],
    "insert_pars" : ["x","y","z"],
    "edit_pars"   : ["id","x","y","z"],
    "topics_pars" : ["id","x","y","z"]
}
```
The part of `nextpages` will be explained below. For *test_b.json*:
```json
{
    "current_table": "test_b",
    "current_key" : "tid",
    "current_id_auto": "tid",
    "insert_pars" : ["id","child"],
    "edit_pars"   : ["tid","child","id"],
    "topics_pars" : ["tid","child","id"]
}
```
With the two files, we can create models and schema. Note that all the CRUD methods are already implemented in `Model`.
```go
    ta, err := NewModel("test_a.json")
    if err != nil { panic(err) }    
    tb, err := NewModel("test_b.json")
    if err != nil { panic(err) }

    models := make(map[string]Navigate)
    models["ta"] = ta
    models["tb"] = tb

    // only "ta" has defined "nextpages", i.e. when run "topics" on "ta", call "tb"'s action "edit" 
    // 
    tt := make(map[string]interface{})
    tt["topics"] = func(args ...url.Values) error {
        return tb.Edit(args...)
    }
    actions := make(map[string]map[string]interface{})
    actions["ta"] = tt

    schema := &Schema{Models:methods, Actions:actions}
    model.Scheme = schema
```
At this step, we have two models `ta`, `tb` and `schema`. Here is the full program to generate APIs of model `ta`:
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

    ta, err := godbi.NewModel("test_a.json")
    if err != nil { panic(err) }
    tb, err := godbi.NewModel("test_b.json")
    if err != nil { panic(err) }

    models := make(map[string]godbi.Navigate)
    models["ta"] = ta
    models["tb"] = tb
    action := make(map[string]interface{})
    action["topics"] = func(args ...url.Values) error { return tb.Topics(args...) }
    action["insert"] = func(args ...url.Values) error { return tb.Insert(args...) }
    action["update"] = func(args ...url.Values) error { return tb.Update(args...) }
    actions := make(map[string]map[string]interface{})
    actions["tb"] = action
    schema := &godbi.Schema{Models:models, Actions:actions}

    ta.Scheme = schema
    tb.Scheme = schema

    db.Exec(`drop table if exists test_a`)
    db.Exec(`CREATE TABLE test_a (id int auto_increment not null primary key,
        x varchar(8), y varchar(8), z varchar(8))`)
    db.Exec(`drop table if exists test_b`)
    db.Exec(`CREATE TABLE test_b (tid int auto_increment not null primary key,
        child varchar(8), id int)`)

    // the 1st web requests is assumed to add id=1 to the test_a and test_b tables:
    //
    hash := url.Values{"x":[]string{"a1234567"},"y":[]string{"b1234567"},"z":[]string{"temp"}, "child":[]string{"john"}}
    ta.UpdateModel(db, hash)
    if err = ta.Insupd(); err != nil { panic(err) }

    // the 2nd request, becaues [x,y] is defined to the unique key in ta, this would't create a new id
    // but will add a new record to tb from Nextpages for id=1, since update triggers its insert too
    // 
    hash = url.Values{"x":[]string{"a1234567"},"y":[]string{"b1234567"},"z":[]string{"zzzzz"}, "child":[]string{"sam"}}
    ta.UpdateModel(db, hash)
    if err = ta.Insupd(); err != nil { panic(err) }

    // the 3rd request adds id=2
    //
    hash = url.Values{"x":[]string{"c1234567"},"y":[]string{"d1234567"},"z":[]string{"e1234"},"child":[]string{"mary"}}
    ta.UpdateModel(db, hash)
    if err = ta.Insupd(); err != nil { panic(err) }

    // the 4th request add id=3
    //
    hash = url.Values{"x":[]string{"e1234567"},"y":[]string{"f1234567"},"z":[]string{"e1234"},"child":[]string{"marcus"}}
    ta.UpdateModel(db, hash)
    if err = ta.Insupd(); err != nil { panic(err) }

    // ready to retrieve 
    err = ta.Topics()
    if err != nil { panic(err) }
    lists := ta.LISTS
    log.Printf("%v", lists)

    os.Exit(0)
}

```
`Model` is used for 
- _InsertPars_ defines column names used for insert a new data
- _InsupdPars_ defines column names used for uniqueness 
- _EditPars_ defines which columns to be returned in *search one* action *Edit*.
- _TopicsPars_ defines which columns to be returned in *search many* action *Topics*.

#### 3.1.2) Incoming data *ARGS*

- Case *search many*, it contains data for *pagination*. 
- Case *insert*, it stores the new row as hash (so the package takes column values of _EditPars_ from *ARGS*).

#### 3.1.3) Output *LISTS*

In case of *search* (*Edit* and *Topics*), the output data are stored in *LISTS*.

#### 3.1.4) Pagination, used in *Topics*.
- *ARGS[SORTBY]* defines sorting by which column
- *ARGS[SORTREVERSE]* defines if a reverse sort
- *ARGS[ROWCOUNT]* defines how many records on each page, an incoming data
- *ARGS[PAGENO]* defines which page number, an incoming data
- *ARGS[TOTALNO]* defines total records available, an output data

Based on those information, developers can build paginations.

#### 3.1.5) Nextpages, calling multiple tables

In many applications, your data involve multiple tables. This function is especially important in TDengine because it's not a relational database and thus has no *JOIN* to use. 

You can define the retrival logic in *Nextpages*, usually expressed as a JSON struct. Assuming there are three *model*s: *testing1*, *testing2* and *testing3*, and you are working in *testing1* now. 
```
"nextpages": {
    "topics" : [
      {"model":"testing2", "action": "topics", "relate_item":{"id":"fid"}},
      {"model":"testing3", "action": "topics"}
    ] ,
    "edit" : [...]
}
```

Thus when you run "topics" on the current model *testing1*, another action "topics" on model "testing2" will be triggered for each returned row. The new action on *testing2* is restricted to have that its column *fid* the same value as *testing1*'s *id*, as in *relate_item*. 

The returned data will be attached to original row under the special key named *testing2_topics*.

- Meanwhile, the above example runs action *topics* on *testing2* once, because there is no *relate_item* in the record.
- The returned will be stored in class variable *OTHER* under key *testing3_topics*.


### 3.2) Create an instance, *NewModel*

An instance is usually created from a JSON string defining the table schema and logic in relationship between tables:
```
model, err := &godbi.Model(json_string)
if err != nil {
        panic(err)
}
// create dbi as above
model.DBI = dbi
// create args as a map
model.ARGS = args

```

If you need to call mutiple tables in one function, you need to put other model instances into *Storage*:
```
// create the database handler "db"
    c := newconf("config.json")
    db, err := sql.Open(c.Db_type, c.Dsn_2)
    if err != nil { panic(err) }
    
// create your current model named "atesting", note that we have nextpages in it 
    model, err := NewModel(`{
    "crud": {
        "current_table": "atesting",
        "current_key" : "id"
        },
    "insupd_pars" : ["x","y"],
    "insert_pars" : ["x","y","z"],
    "edit_pars" : ["x","y","z","id"],
    "topics_pars" : ["id","x","y","z"],
    "nextpages" : {
        "topics" : [
            {"model":"testing", "action":"topics", "relate_item":{"id":"id"}}
        ]
    }
}`)
    if err != nil { panic(err) }
    model.Db = db
    model.ARGS  = make(map[string]interface{})
    model.OTHER = make(map[string]interface{})

// create another model with name "testing"
    st, err := NewModel(`{
    "crud": {
        "current_table": "testing",
        "current_key" : "tid"
    },
    "insert_pars" : ["id","child"],
    "edit_pars"   : ["tid","child","id"],
    "topics_pars" : ["tid","child","id"]
}`)
    if err != nil { panic(err) }
    st.Db = db
    st.ARGS  = make(map[string]interface{})
    st.OTHER = make(map[string]interface{})

// create a storage to mark "testing"
    storage := make(map[string]map[string]interface{})
    storage["model"]= make(map[string]interface{})
    storage["model"]["testing"]= st
    storage["action"]= make(map[string]interface{})
    tt := make(map[string]interface{})
    tt["topics"] = func(args ...map[string]interface{}) error {
        return st.Topics(args...)
    }
    storage["action"]["testing"] = tt
```

### 3.3) Actions (functions) on *Model*

#### 3.3.1) Insert one row, *Insert*
```
err = model.Insert()
```
It will takes values from *ARGS* using pre-defined column names in *InsertPars*. If you miss the primary key, the package will automatically assign *now* to be the value.


#### 3.3.2) Insert or Retrieve an old row, *Insupd*

You insert one row as in *Insert* but if it is already in the table, retrieve it.
```
err = model.Insupd()
```
It identifies the uniqueness by the combined column valuse defined in *InsupdPars*. In both the cases, you get the ID in *model.LastID*, the row in *CurrentRow*, and the case in *Updated* (true for old record, and false for new). 


#### 3.3.3) Select many rows, *Topics*

Search many by *Topics*:
```
restriction := map[string]interface{}{"len":10}
err = crud.Topics(restriction)
```
which returns all records with columns defined in *TopicsPars* with restriction *len=10*. The returned data is in *model.LISTS*

Since you have assigned *nextpages* for this action, for each row it retrieve, another *Topics* on model *testing* will run using the constraint that *id* in *testing* should take the value in the original row.


#### 3.3.4) Select one row, *Edit*

```
err = crud.EditHash()
```
Here you select by its primary key value (the timestamp), which is assumed to be in *ARGS*. The returned data is in *model.LISTS*. Optionally, you may put a restriction. 


#### 3.3.5) Sort order, *OrderString*

This returns you the sort string used in the *select many*. If you inherit this class you can override this function to use your own sorting logic.


## SAMPLES

Please check those test files:

- DBI: [dbi_test.go](https://github.com/genelet/godbi/blob/master/dbi_test.go)
- Crud: [crud_test.go](https://github.com/genelet/godbi/blob/master/crud_test.go)
- Model: [model_test.go](https://github.com/genelet/godbi/blob/master/model_test.go)









