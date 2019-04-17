csvutil [![GoDoc](https://godoc.org/github.com/jszwec/csvutil?status.svg)](http://godoc.org/github.com/jszwec/csvutil) [![Build Status](https://travis-ci.org/jszwec/csvutil.svg?branch=master)](https://travis-ci.org/jszwec/csvutil) [![Build status](https://ci.appveyor.com/api/projects/status/eiyx0htjrieoo821/branch/master?svg=true)](https://ci.appveyor.com/project/jszwec/csvutil/branch/master) [![Go Report Card](https://goreportcard.com/badge/github.com/jszwec/csvutil)](https://goreportcard.com/report/github.com/jszwec/csvutil) [![codecov](https://codecov.io/gh/jszwec/csvutil/branch/master/graph/badge.svg)](https://codecov.io/gh/jszwec/csvutil)
=================

<p align="center">
  <img style="float: right;" src="https://user-images.githubusercontent.com/3941256/33054906-52b4bc08-ce4a-11e7-9651-b70c5a47c921.png"/ width=200>
</p>

Package csvutil provides fast and idiomatic mapping between CSV and Go (golang) values.

This package does not provide a CSV parser itself, it is based on the [Reader](https://godoc.org/github.com/jszwec/csvutil#Reader) and [Writer](https://godoc.org/github.com/jszwec/csvutil#Writer)
interfaces which are implemented by eg. std Go (golang) [csv package](https://golang.org/pkg/encoding/csv). This gives a possibility
of choosing any other CSV writer or reader which may be more performant.

Installation
------------

    go get github.com/jszwec/csvutil

Requirements
-------------

* Go1.7+

Index
------

1. [Examples](#examples)
	1. [Unmarshal](#examples_unmarshal)
	2. [Marshal](#examples_marshal)
	3. [Unmarshal and metadata](#examples_unmarshal_and_metadata)
	4. [But my CSV file has no header...](#examples_but_my_csv_has_no_header)
	5. [Decoder.Map - data normalization](#examples_decoder_map)
	6. [Different separator/delimiter](#examples_different_separator)
	7. [Decoder and interface values](#examples_decoder_interface_values)
	8. [Custom time.Time format](#examples_time_format)
	9. [Custom struct tags](#examples_struct_tags)
2. [Performance](#performance)
	1. [Unmarshal](#performance_unmarshal)
	2. [Marshal](#performance_marshal)

Example <a name="examples"></a>
--------

### Unmarshal <a name="examples_unmarshal"></a>

Nice and easy Unmarshal is using the Go std [csv.Reader](https://golang.org/pkg/encoding/csv/#Reader) with its default options. Use [Decoder](https://godoc.org/github.com/jszwec/csvutil#Decoder) for streaming and more advanced use cases.

```go
	var csvInput = []byte(`
name,age,CreatedAt
jacek,26,2012-04-01T15:00:00Z
john,,0001-01-01T00:00:00Z`,
	)

	type User struct {
		Name      string `csv:"name"`
		Age       int    `csv:"age,omitempty"`
		CreatedAt time.Time
	}

	var users []User
	if err := csvutil.Unmarshal(csvInput, &users); err != nil {
		fmt.Println("error:", err)
	}

	for _, u := range users {
		fmt.Printf("%+v\n", u)
	}

	// Output:
	// {Name:jacek Age:26 CreatedAt:2012-04-01 15:00:00 +0000 UTC}
	// {Name:john Age:0 CreatedAt:0001-01-01 00:00:00 +0000 UTC}
```

### Marshal <a name="examples_marshal"></a>

Marshal is using the Go std [csv.Writer](https://golang.org/pkg/encoding/csv/#Writer) with its default options. Use [Encoder](https://godoc.org/github.com/jszwec/csvutil#Encoder) for streaming or to use a different Writer.

```go
	type Address struct {
		City    string
		Country string
	}

	type User struct {
		Name string
		Address
		Age       int `csv:"age,omitempty"`
		CreatedAt time.Time
	}

	users := []User{
		{
			Name:      "John",
			Address:   Address{"Boston", "USA"},
			Age:       26,
			CreatedAt: time.Date(2010, 6, 2, 12, 0, 0, 0, time.UTC),
		},
		{
			Name:    "Alice",
			Address: Address{"SF", "USA"},
		},
	}

	b, err := csvutil.Marshal(users)
	if err != nil {
		fmt.Println("error:", err)
	}
	fmt.Println(string(b))

	// Output:
	// Name,City,Country,age,CreatedAt
	// John,Boston,USA,26,2010-06-02T12:00:00Z
	// Alice,SF,USA,,0001-01-01T00:00:00Z
```

### Unmarshal and metadata <a name="examples_unmarshal_and_metadata"></a>

It may happen that your CSV input will not always have the same header. In addition
to your base fields you may get extra metadata that you would still like to store.
[Decoder](https://godoc.org/github.com/jszwec/csvutil#Decoder) provides 
[Unused](https://godoc.org/github.com/jszwec/csvutil#Decoder.Unused) method, which after each call to 
[Decode](https://godoc.org/github.com/jszwec/csvutil#Decoder.Decode) can report which header indexes 
were not used during decoding. Based on that, it is possible to handle and store all these extra values.

```go
	type User struct {
		Name      string            `csv:"name"`
		City      string            `csv:"city"`
		Age       int               `csv:"age"`
		OtherData map[string]string `csv:"-"`
	}

	csvReader := csv.NewReader(strings.NewReader(`
name,age,city,zip
alice,25,la,90005
bob,30,ny,10005`))

	dec, err := csvutil.NewDecoder(csvReader)
	if err != nil {
		log.Fatal(err)
	}

	header := dec.Header()
	var users []User
	for {
		u := User{OtherData: make(map[string]string)}

		if err := dec.Decode(&u); err == io.EOF {
			break
		} else if err != nil {
			log.Fatal(err)
		}

		for _, i := range dec.Unused() {
			u.OtherData[header[i]] = dec.Record()[i]
		}
		users = append(users, u)
	}

	fmt.Println(users)

	// Output:
	// [{alice la 25 map[zip:90005]} {bob ny 30 map[zip:10005]}]
```

### But my CSV file has no header... <a name="examples_but_my_csv_has_no_header"></a>

Some CSV files have no header, but if you know how it should look like, it is
possible to define a struct and generate it. All that is left to do, is to pass
it to a decoder.

```go
	type User struct {
		ID   int
		Name string
		Age  int `csv:",omitempty"`
		City string
	}

	csvReader := csv.NewReader(strings.NewReader(`
1,John,27,la
2,Bob,,ny`))

	// in real application this should be done once in init function.
	userHeader, err := csvutil.Header(User{}, "csv")
	if err != nil {
		log.Fatal(err)
	}

	dec, err := csvutil.NewDecoder(csvReader, userHeader...)
	if err != nil {
		log.Fatal(err)
	}

	var users []User
	for {
		var u User
		if err := dec.Decode(&u); err == io.EOF {
			break
		} else if err != nil {
			log.Fatal(err)
		}
		users = append(users, u)
	}

	fmt.Printf("%+v", users)

	// Output:
	// [{ID:1 Name:John Age:27 City:la} {ID:2 Name:Bob Age:0 City:ny}]
```

### Decoder.Map - data normalization <a name="examples_decoder_map"></a>

The Decoder's [Map](https://godoc.org/github.com/jszwec/csvutil#Decoder.Map) function is a powerful tool that can help clean up or normalize
the incoming data before the actual decoding takes place.

Lets say we want to decode some floats and the csv input contains some NaN values, but these values are represented by the 'n/a' string. An attempt to decode 'n/a' into float will end up with error, because strconv.ParseFloat expects 'NaN'. Knowing that, we can implement a Map function that will normalize our 'n/a' string and turn it to 'NaN' only for float types.

```go
	dec, err := NewDecoder(r)
	if err != nil {
		log.Fatal(err)
	}

	dec.Map = func(field, column string, v interface{}) string {
		if _, ok := v.(float64); ok && field == "n/a" {
			return "NaN"
		}
		return field
	}
```

Now our float64 fields will be decoded properly into NaN. What about float32, float type aliases and other NaN formats? Look at the full example [here](https://gist.github.com/jszwec/2bb94f8f3612e0162eb16003701f727e).

### Different separator/delimiter <a name="examples_different_separator"></a>

Some files may use different value separators, for example TSV files would use `\t`. The following examples show how to set up a Decoder and Encoder for such use case.

#### Decoder:
```go
	csvReader := csv.NewReader(r)
	csvReader.Comma = '\t'

	dec, err := NewDecoder(csvReader)
	if err != nil {
		log.Fatal(err)
	}

	var users []User
	for {
		var u User
		if err := dec.Decode(&u); err == io.EOF {
			break
		} else if err != nil {
			log.Fatal(err)
		}
		users = append(users, u)
	}

```

#### Encoder:
```go
	var buf bytes.Buffer

	w := csv.NewWriter(&buf)
        w.Comma = '\t'
	enc := csvutil.NewEncoder(w)

	for _, u := range users {
		if err := enc.Encode(u); err != nil {
			log.Fatal(err)
		}
        }

	w.Flush()
	if err := w.Error(); err != nil {
		log.Fatal(err)
	}
```

### Decoder and interface values <a name="examples_decoder_interface_values"></a>

In the case of interface struct fields data is decoded into strings. However, if Decoder finds out that
these fields were initialized with pointer values of a specific type prior to decoding, it will try to decode data into that type.

Why only pointer values? Because these values must be both addressable and settable, otherwise Decoder
will have to initialize these types on its own, which could result in losing some unexported information.

If interface stores a non-pointer value it will be replaced with a string.

This example will show how this feature could be useful:
```go
package main

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"io"
	"log"

	"github.com/jszwec/csvutil"
)

// Value defines one record in the csv input. In this example it is important
// that Type field is defined before Value. Decoder reads headers and values
// in the same order as struct fields are defined.
type Value struct {
	Type  string      `csv:"type"`
	Value interface{} `csv:"value"`
}

func main() {
	// lets say our csv input defines variables with their types and values.
	data := []byte(`
type,value
string,string_value
int,10
`)

	dec, err := csvutil.NewDecoder(csv.NewReader(bytes.NewReader(data)))
	if err != nil {
		log.Fatal(err)
	}

	// we would like to read every variable and store their already parsed values
	// in the interface field. We can use Decoder.Map function to initialize
	// interface with proper values depending on the input.
	var value Value
	dec.Map = func(field, column string, v interface{}) string {
		if column == "type" {
			switch field {
			case "int": // csv input tells us that this variable contains an int.
				var n int
				value.Value = &n // lets initialize interface with an initialized int pointer.
			default:
				return field
			}
		}
		return field
	}

	for {
		value = Value{}
		if err := dec.Decode(&value); err == io.EOF {
			break
		} else if err != nil {
			log.Fatal(err)
		}

		if value.Type == "int" {
			// our variable type is int, Map func already initialized our interface
			// as int pointer, so we can safely cast it and use it.
			n, ok := value.Value.(*int)
			if !ok {
				log.Fatal("expected value to be *int")
			}
			fmt.Printf("value_type: %s; value: (%T) %d\n", value.Type, value.Value, *n)
		} else {
			fmt.Printf("value_type: %s; value: (%T) %v\n", value.Type, value.Value, value.Value)
		}
	}

	// Output:
	// value_type: string; value: (string) string_value
	// value_type: int; value: (*int) 10
}
```

### Custom time.Time format <a name="examples_time_format"></a>

Type [time.Time](https://golang.org/pkg/time/#Time) can be used as is in the struct fields by both Decoder and Encoder
due to the fact that both have builtin support for [encoding.TextUnmarshaler](https://golang.org/pkg/encoding/#TextUnmarshaler) and [encoding.TextMarshaler](https://golang.org/pkg/encoding/#TextMarshaler). This means that by default
Time has a specific format; look at [MarshalText](https://golang.org/pkg/time/#Time.MarshalText) and [UnmarshalText](https://golang.org/pkg/time/#Time.UnmarshalText). This example shows how to override it.
```go
type Time struct {
	time.Time
}

const format = "2006/01/02 15:04:05"

func (t Time) MarshalCSV() ([]byte, error) {
	var b [len(format)]byte
	return t.AppendFormat(b[:0], format), nil
}

func (t *Time) UnmarshalCSV(data []byte) error {
	tt, err := time.Parse(format, string(data))
	if err != nil {
		return err
	}
	*t = Time{Time: tt}
	return nil
}
```

### Custom struct tags <a name="examples_struct_tags"></a>

Like in other Go encoding packages struct field tags can be used to set
custom names or options. By default encoders and decoders are looking at `csv` tag.
However, this can be overriden by manually setting the Tag field.

```go
	type Foo struct {
		Bar int `custom:"bar"`
	}
```

```go
	dec, err := csvutil.NewDecoder(r)
	if err != nil {
		log.Fatal(err)
	}
	dec.Tag = "custom"
```

```go
	enc := csvutil.NewEncoder(w)
	enc.Tag = "custom"
```

Performance
------------

csvutil provides the best encoding and decoding performance with small memory usage.

### Unmarshal <a name="performance_unmarshal"></a>

[benchmark code](https://gist.github.com/jszwec/e8515e741190454fa3494bcd3e1f100f)

#### csvutil:
```
BenchmarkUnmarshal/csvutil.Unmarshal/1_record-8         	  300000	      5852 ns/op	    6900 B/op	      32 allocs/op
BenchmarkUnmarshal/csvutil.Unmarshal/10_records-8       	  100000	     13946 ns/op	    7924 B/op	      41 allocs/op
BenchmarkUnmarshal/csvutil.Unmarshal/100_records-8      	   20000	     95234 ns/op	   18100 B/op	     131 allocs/op
BenchmarkUnmarshal/csvutil.Unmarshal/1000_records-8     	    2000	    903502 ns/op	  120652 B/op	    1031 allocs/op
BenchmarkUnmarshal/csvutil.Unmarshal/10000_records-8    	     200	   9273741 ns/op	 1134694 B/op	   10031 allocs/op
BenchmarkUnmarshal/csvutil.Unmarshal/100000_records-8   	      20	  94125839 ns/op	11628908 B/op	  100031 allocs/op
```

#### gocsv:
```
BenchmarkUnmarshal/gocsv.Unmarshal/1_record-8           	  200000	     10363 ns/op	    7651 B/op	      96 allocs/op
BenchmarkUnmarshal/gocsv.Unmarshal/10_records-8         	   50000	     31308 ns/op	   13747 B/op	     306 allocs/op
BenchmarkUnmarshal/gocsv.Unmarshal/100_records-8        	   10000	    237417 ns/op	   72499 B/op	    2379 allocs/op
BenchmarkUnmarshal/gocsv.Unmarshal/1000_records-8       	     500	   2264064 ns/op	  650135 B/op	   23082 allocs/op
BenchmarkUnmarshal/gocsv.Unmarshal/10000_records-8      	      50	  24189980 ns/op	 7023592 B/op	  230091 allocs/op
BenchmarkUnmarshal/gocsv.Unmarshal/100000_records-8     	       5	 264797120 ns/op	75483184 B/op	 2300104 allocs/op
```

#### easycsv:
```
BenchmarkUnmarshal/easycsv.ReadAll/1_record-8           	  100000	     13287 ns/op	    8855 B/op	      81 allocs/op
BenchmarkUnmarshal/easycsv.ReadAll/10_records-8         	   20000	     66767 ns/op	   24072 B/op	     391 allocs/op
BenchmarkUnmarshal/easycsv.ReadAll/100_records-8        	    3000	    586222 ns/op	  170537 B/op	    3454 allocs/op
BenchmarkUnmarshal/easycsv.ReadAll/1000_records-8       	     300	   5630293 ns/op	 1595662 B/op	   34057 allocs/op
BenchmarkUnmarshal/easycsv.ReadAll/10000_records-8      	      20	  60513920 ns/op	18870410 B/op	  340068 allocs/op
BenchmarkUnmarshal/easycsv.ReadAll/100000_records-8     	       2	 623618489 ns/op	190822456 B/op	 3400084 allocs/op
```

### Marshal <a name="performance_marshal"></a>

[benchmark code](https://gist.github.com/jszwec/31980321e1852ebb5615a44ccf374f17)

#### csvutil:
```
BenchmarkMarshal/csvutil.Marshal/1_record-8               200000              6542 ns/op            9568 B/op         11 allocs/op
BenchmarkMarshal/csvutil.Marshal/10_records-8             100000             21458 ns/op           10480 B/op         21 allocs/op
BenchmarkMarshal/csvutil.Marshal/100_records-8             10000            167195 ns/op           27890 B/op        112 allocs/op
BenchmarkMarshal/csvutil.Marshal/1000_records-8             1000           1619843 ns/op          168210 B/op       1014 allocs/op
BenchmarkMarshal/csvutil.Marshal/10000_records-8             100          16190060 ns/op         1525812 B/op      10017 allocs/op
BenchmarkMarshal/csvutil.Marshal/100000_records-8             10         163375841 ns/op        22369524 B/op     100021 allocs/op
```

#### gocsv:
```
BenchmarkMarshal/gocsv.Marshal/1_record-8           	  200000	      7202 ns/op	    5922 B/op	      83 allocs/op
BenchmarkMarshal/gocsv.Marshal/10_records-8         	   50000	     31821 ns/op	    9427 B/op	     390 allocs/op
BenchmarkMarshal/gocsv.Marshal/100_records-8        	    5000	    285885 ns/op	   52773 B/op	    3451 allocs/op
BenchmarkMarshal/gocsv.Marshal/1000_records-8       	     500	   2806405 ns/op	  452517 B/op	   34053 allocs/op
BenchmarkMarshal/gocsv.Marshal/10000_records-8      	      50	  28682052 ns/op	 4412157 B/op	  340065 allocs/op
BenchmarkMarshal/gocsv.Marshal/100000_records-8     	       5	 286836492 ns/op	51969227 B/op	 3400083 allocs/op
```
