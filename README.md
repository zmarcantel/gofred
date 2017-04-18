gofred [![Build Status](https://travis-ci.org/zmarcantel/gofred.svg?branch=master)](https://travis-ci.org/zmarcantel/gofred)
======

example
=======

First, we create a client

```go
client, err := gofred.NewClient(MY_API_KEY, gofred.JSON) // use json for responses
```

categories
----------

#### single

To get information about a single category with a known category ID:

```go
// 125 = Trade Balance
category, err := client.Category(125)
```

#### children

To get information about a category's children (if any):

```go
// 13 = U.S. Trade & International Transactions
category, err := client.CategoryChildren(13, time.Unix(0, 0), time.Now())
```

#### related

To get information about all categories (if any) related to a given category:

```go
// 32073 = States in the St. Louis FED District
category, err := client.RelatedCategories(32073, time.Unix(0, 0), time.Now())
```


testing
=======

The API requires an API key.

An API key can be registered: [here](http://api.stlouisfed.org/api_key.html).

In order to run tests, some file in the package must have the constant `API_KEY` defined in it.

For `travis-ci` this is generated using their file-decryption method.
