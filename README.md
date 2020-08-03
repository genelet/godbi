# godbi
_godbi_ adds a set of high-level functions to the official SQL handle in GO, for easier database executions and queries. Check *godoc* from [here](https://godoc.org/github.com/genelet/godbi) for definitions.

[![GoDoc](https://godoc.org/github.com/genelet/godbi?status.svg)](https://godoc.org/github.com/genelet/godbi)

There are three levels of usages:
- _Basic_: operating on raw SQL statements and stored procedures, and receiving data as a _slice_ of rows. Each row is represented as a _map_ between the column names (_string_) and column value (_interface{}_).
- _Table_: operating on a specific table and fulfilling the CRUD operations using _map_ data.
- _Advanced_: operating on multiple tables, called Models, as in MVC pattern in web applications; and fulfilling the RESTful and GraphQL actions.

Note that all functions are consistently returning _error_ if failed, or _nil_ if succeeded.

_godbi_ is an ideal replacement of ORM. It achieves the common SQL, CRUD, RESTful and GraphQL tasks very easily and efficiently.
The package is fully tested in MySQL and PostgreSQL, and is assumed to work with other relational databases.


### Installation

```
$ go get -u github.com/genelet/godbi
```
<!-- go mod init github.com/genelet/godbi -->



## Chapter 1. BASIC USAGE


### 1.1) Type _DBI_

The _DBI_ type simply embeds the standard SQL handle.
```
package godbi

type DBI struct {
    *sql.DB          // Note this is the pointer to the handle
    LastId    int64  // read only, saves the last inserted id
    Affected  int64  // read only, saves the affected rows
}

```

#### Create a new handle

```
dbi := &DBI{DB: the_standard_sql_handle}
```

#### Example

In this example, we create a MySQL handle using database credentials in the environment; then create a new table _letters_ and add 3 rows. We query the data using _SelectSQL_ and put the result into _lists_ as slice of maps.
```
package main

    dbUser := os.Getenv("DBUSER")
    dbPass := os.Getenv("DBPASS")
    dbName := os.Getenv("DBNAME")
    db, err := sql.Open("mysql", dbUser + ":" + dbPass + "@/" + dbName)
    if err != nil { panic(err) }
    defer db.Close()

    dbi := &godbi.DBI{DB:db}

    // create a new table and insert some data using ExecSQL
    //
    err = dbi.ExecSQL(`DROP TABLE IF EXISTS letters`)
    if err != nil { panic(err) }
    err = dbi.ExecSQL(`CREATE TABLE letters (
        id int auto_increment primary key,
        x varchar(1))`)
    if err != nil { panic(err) }
    err = dbi.ExecSQL(`INSERT INTO letters (x) VALUES ('m')`)
    if err != nil { panic(err) }
    err = dbi.ExecSQL(`INSERT INTO letters (x) VALUES ('n')`)
    if err != nil { panic(err) }
    err = dbi.ExecSQL(`INSERT INTO letters (x) VALUES ('p')`)
    if err != nil { panic(err) }

    // select data from the table and put them into lists
    lists := make([]map[string]interface{},0)
    sql := "SELECT id, x FROM letters"
    err = dbi.SelectSQL(&lists, sql)
    if err != nil { panic(err) }

    // print it
    log.Printf("%v", lists)

    dbi.ExecSQL(`DROP TABLE letters`)

    os.Exit(0)
}
```
Running this example will report something like
```
[map[id:1 x:m] map[id:2 x:n] map[id:3 x:p]]
```


### 1.2) Execution with _ExecSQL_ & _DoSQL_

Definition:
```
func (*DBI) ExecSQL(query string, args ...interface{}) error
func (*DBI) DoSQL  (query string, args ...interface{}) error
```
Similar to SQL's _Exec_, these functions execute _INSERT_ or _UPDATE_ queries. The returned value should be checked to assert if the executions are successful.

The difference between the two functions is that _DoSQL_ runs a prepared statement and is safe for concurrent use by multiple goroutines.


### 1.3) Queries with _SELECT_ 

#### 1.3.1)  *QuerySQL* & *SelectSQL*

Definition:
```
func (*DBI) QuerySQL (lists *[]map[string]interface{}, query string, args ...interface{}) error
func (*DBI) SelectSQL(lists *[]map[string]interface{}, query string, args ...interface{}) error
```
Run query and put the result into *lists*. The data types are determined automatically by the generic SQL handle. For example:
```
lists := make([]map[string]interface{})
err = dbi.QuerySQL(&lists,
    `SELECT ts, id, name, len, flag, fv FROM mytable WHERE id=?`, 1234)
```
This will select all rows with _id=1234_.
```
[
    {"ts":"2019-12-15 01:01:01", "id":1234, "name":"company", "len":30, "flag":true, "fv":789.123},
    ....
]
```
The difference between the two functions is that _SelectSQL_ runs a prepared statement.

#### 1.3.2) *QuerySQLType* & *SelectSQLType*

Definition:
```
func (*DBI) QuerySQLType (lists *[]map[string]interface{}, typeLabels []string, query string, args ...interface{}) error
func (*DBI) SelectSQLType(lists *[]map[string]interface{}, typeLabels []string, query string, args ...interface{}) error
```
They differ from the above *QuerySQL* by specifying data types to the rows. While the generic handle could correctly figure out the types in most cases, occasionally it fails because there is no exact matching between SQL typies to GOLANG data types.

Here is an example. We have assign _string_, _int_, _string_, _int8_, _bool_ and _float32_ to the corresponding columns:
```
err = dbi.QuerySQLType(&lists, []string{"string", "int", "string", "int8", "bool", "float32},
    `SELECT ts, id, name, len, flag, fv FROM mytable WHERE id=?`, 1234)
```

#### 1.3.3) *QuerySQLLabel* & *SelectSQLLable*

Definition:
```
func (*DBI) QuerySQLLabel (lists *[]map[string]interface{}, selectLabels []string, query string, args ...interface{}) error
func (*DBI) SelectSQLLabel(lists *[]map[string]interface{}, selectLabels []string, query string, args ...interface{}) error
```
They differ from the above *QuerySQL* by specifying map keys, called _selectLabels_, instead of using the default column names. For example:
```
lists := make([]map[string]interface{})
err = dbi.QuerySQLLabel(&lists, []string{"time stamp", "record ID", "recorder name", "length", "flag", "values"},
    `SELECT ts, id, name, len, flag, fv FROM mytable WHERE id=?`, 1234)
```
So the result will show the _labels_ as the map keys:
```
[
    {"time stamp":"2019-12-15 01:01:01", "record ID":1234, "recorder name":"company", "length":30, "flag":true, "values":789.123},
    ....
]
```

#### 1.3.4) *QuerySQLTypeLabel* & *SelectSQlTypeLabel*

Definition:
```
func (*DBI) QuerySQLTypeLabel (lists *[]map[string]interface{}, typeLabels []string, selectLabels []string, query string, args ...interface{}) error
func (*DBI) SelectSQLTypeLabel(lists *[]map[string]interface{}, typeLabels []string, selectLabels []string, query string, args ...interface{}) error
```
These functions assign both data types and column names in the queries.


### 1.4) Query one-row data with _SELECT_ 

In some cases we may just want to select one row from a query. 

#### 1.4.1) *GetSQLLable*

Definition:
```
func (*DBI) GetSQLLabel(res map[string]interface{}, query string, selectLabels []string, args ...interface{}) error
```
which is similar to *SelectSQLLable* but has only one row output to *res*.

#### 1.4.2) *GetArgs*

Definition:
```
func (*DBI) GetArgs(res url.Values, query string, args ...interface{}) error
```
which is similar to *SelectSQL* but has only one row output to *res* of of type [url.Values](https://golang.org/pkg/net/url/). This function will be used mainly in web applications, where HTTP request data are expressed in _url.Values_.


### 1.5) Stored Procedure

_godbi_ runs stored procedures easily as well.

#### 1.5.1) *DoProc*

Definition:
```
func (*DBI) DoProc(res map[string]interface{}, names []string, proc_name string, args ...interface{}) error
```
It runs a stored procedure *proc_name* with IN data in *args*. The OUT data will be placed in the map *res* using keys defined in *names*. Note that the output columns should have already been defined in *proc_name*.

If the procedure has no output data to receive, just assign *names* to be _nil_.

#### 1.5.2) *SelectDoProc*

Definition:
```
func (*DBI) SelectDoProc(lists *[]map[string]interface{}, res map[string]interface{}, names []string, proc_name string, args ...interface{}) error
```
Similar to *DoProc* but it receives _SELECT_ data into *lists*, providing *proc_name* contains such a query. 

Here is a full example.
```
ackage main

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

    dbi.ExecSQL(`drop procedure if exists proc_w2`)
    dbi.ExecSQL(`drop table if exists letters`)
    err = dbi.ExecSQL(`create table letters(id int auto_increment primary key, x varchar(1))`)
    if err != nil { panic(err) }

    err = dbi.ExecSQL(`create procedure proc_w2(IN x0 varchar(1),OUT y0 int)
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



## Chapter 2. TABLE USAGE


Somet times it is more convenient to work on a table by using  
```
type Crud struct {
    DBI
    CurrentTable string    // the current table name 
    CurrentTables []*Table // optional, for read-all SELECT with other joined tables 
    CurrentKey string      // the single primary key of the table    
    CurrentKeys []string   // optional, if the primary key has multiple columns   
    Updated bool           // for Insupd() only, if the row is updated or new
}
```
where *CurrentTable* is the table you are working with, *CurrentKey* the primary key in the table, always a [*timestamp*](https://www.taosdata.com/en/documentation/taos-sql/#Data-Query). *LastID*, *CurrentRow* and *Updated* are for the last inserted row.

You create an instance of *Crud* by
```
crud := &godbi.Crud{Db:db, CurrentTable:mytable, CurrentKey:ts}
```


### 2.1) Insert one row, *InsertHash*
```
err = crud.InsertHash(map[string]interface{}{
        {"ts":"2019-12-31 23:59:59.9999", "id":7890, "name":"last day", "fv":123.456}
})
```
If you miss the primary key, the package will automatically assign *now* to be the value.


### 2.2) Insert or Retrieve an old row, *InsupdHash*

Sometimes a record may already be existing in the table, so you'd like to insert if it is not there, or retrieve it. Function *InsupdHash* is for this purpose:
```
err = crud.InsupdHash(map[string]interface{}{
        {"ts":"2019-12-31 23:59:59.9999", "id":7890, "name":"last day", "fv":123.456}},
        []string{"id","name"},
)
```
It identifies the uniqueness by the combined valuse of *id* and *name*. In both the cases, you get the ID in *crud.LastID*, the row in *CurrentRow*, and the case in *Updated* (true for old record, and false for new). 


### 2.3) Select many rows, *TopicsHash*

Search many by *TopicsHash*:
```
lists := make([]map[string]interface{})
restriction := map[string]interface{}{"len":10}
err = crud.TopicsHash(&lists, []string{"ts", "name", "id"}, restriction)
```
which returns all records with restriction *len=10*. You specifically define which columns to return in second argument,
which are *ts*, *name* and *id* here. 

Only three types of _restriction_ are supported in map:
- _key:value_  The *key* has *value*.
- _key:slice_  The *key* has one of values in *slice*.
- _"_gsql":"row sql statement"_  Use the special key *_gsql* to write a raw SQL statment.


### 2.4) Select one row, *EditHash*

```
lists := make([]map[string]interface{})
err = crud.EditHash(&lists, []string{"ts", "name", "id"}, "2019-12-31 23:59:59.9999")
```
Here you select by its primary key value (the timestamp). 

Optionally, you may input an array of few key values and get them all in *lists*. Or you may put a restriction map too.




## Chapter 3. ADVANCED USAGE

*Model* is even a more detailed class operation on TDengine table.

### 3.1) Class *Model*

```
type Model struct {
    Crud `json:"crud,omitempty"`

    ARGS  map[string]interface{}   `json:"args,omitempty"`
    LISTS []map[string]interface{} `json:"lists,omitempty"`
    OTHER map[string]interface{}   `json:"other,omitempty"`

    SORTBY      string `json:"sortby,omitempty"`
    SORTREVERSE string `json:"sortreverse,omitempty"`
    PAGENO      string `json:"pageno,omitempty"`
    ROWCOUNT    string `json:"rowcount,omitempty"`
    TOTALNO     string `json:"totalno,omitempty"`

    Nextpages map[string][]map[string]interface{} `json:"nextpages,omitempty"`
    Storage   map[string]map[string]interface{}   `json:"storage,omitempty"`

    InsertPars []string `json:"insert_pars,omitempty"`
    InsupdPars []string `json:"insupd_Pars,omitempty"`

    EditPars   []string          `json:"edit_pars,omitempty"`
    TopicsPars []string          `json:"topics_pars,omitempty"`
    EditMap    map[string]string `json:"edit_map,omitempty"`
    TopicsMap  map[string]string `json:"topics_map,omitempty"`

    TotalForce int `json:"total_force,omitempty"`
}

```

####  3.1.1) Table column names
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









