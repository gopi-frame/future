# Future - a simple encapsulation of asynchronous operations

## Installation
```
go get -u -v github.com/gopi-frame/future
```

## Import
```
import "github.com/gopi-frame/future"
```

## Usage

### Quick Start

#### Async

```go
package main

import (
	"fmt"
	"github.com/gopi-frame/future"
	"log"
	"net/http"
)

func main() {
	var resp = future.Async[*http.Response](func() *http.Response {
		resp, err := http.Get("https://example.com")
		if err != nil {
			panic(err)
		}
		return resp
	}).CatchAll(func(err error) {
		log.Fatalln(err.Error())
	}).Await()
	fmt.Println(resp.StatusCode)
}
```

#### Value

```go
package main

import (
	"fmt"
	"github.com/gopi-frame/future"
)

func main() {
	value := future.Value("Hello world").Await()
	fmt.Println(value) // Hello world
}
```

#### Timeout

```go
package main

import (
	"fmt"
	"github.com/gopi-frame/exception"
	"github.com/gopi-frame/future"
	"log"
	"net/http"
	"time"
)

func main() {
	resp := future.Timeout[*http.Response](func() *http.Response {
		resp, err := http.Get("https://example.com")
		if err != nil {
			panic(err)
		}
		return resp
	}, time.Second*10).Catch(new(exception.TimeoutException), func(err error) {
		log.Fatalln("timeout")
	}).CatchAll(func(err error) {
		log.Fatalln(err.Error())
	}).Await()
	fmt.Println(resp.StatusCode)
}
```

#### Delay

```golang
package main

import (
	"fmt"
	"github.com/gopi-frame/future"
	"log"
	"net/http"
	"time"
)

func main() {
	resp := future.Delay[*http.Response](func() *http.Response {
		resp, err := http.Get("https://example.com")
		if err != nil {
			panic(err)
		}
		return resp
	}, time.Second*5).CatchAll(func(err error) {
		log.Fatalln(err.Error())
	}).Await()
	fmt.Println(resp.StatusCode)
}
```

#### Foreach

```go
package main

import (
	"github.com/gopi-frame/future"
	"strconv"
)

func main() {
	var numbers = []int{1, 2, 3, 4, 5}
	future.Foreach[int, string](numbers, func(i int) *future.Future[string] {
		return future.Value[string](strconv.Itoa(i))
	}).Await()
}
```

#### Wait

```go
package main

import (
	"fmt"
	"github.com/gopi-frame/future"
	"strings"
)

func main() {
	var numbers = []int{1, 2, 3, 4, 5}
	var futures []*future.Future[int]
	for _, number := range numbers {
		value := number * number
		futures = append(futures, future.Value[int](value))
	}
	results := future.Wait[int](futures...).Await()
	fmt.Println(strings.Join(results.ToArray(), ", ")) // 1, 4, 9, 16, 25
}
```

### Chain Operation

```go
package main

import (
	"fmt"
	"github.com/gopi-frame/future"
)

func main() {
	str := future.Async(func() string {
		return "world"
	}).Then(func(value string) string {
		return fmt.Sprintf("Hello %s", value)
	}, func(err error) {}).Await()
	fmt.Println(str) // Hello world
}
```

### Error Handling

#### Then With OnError
```go
package main

import (
	"encoding/json"
	"fmt"
	"github.com/gopi-frame/future"
	"io"
	"net/http"
)

type RequestError struct{}

func (r *RequestError) Error() string {
	return "request error"
}

type JSONDecodeError struct{}

func (j *JSONDecodeError) Error() string {
	return "json decode error"
}

func main() {
	future.Async(func() map[string]any {
		resp, err := http.Get("https://example.com")
		if err != nil {
			panic(new(RequestError))
		}
		content, err := io.ReadAll(resp.Body)
		if err != nil {
			panic(err)
		}
		defer func() {
			if err := resp.Body.Close(); err != nil {
				panic(err)
			}
		}()
		var result = make(map[string]any)
		if err := json.Unmarshal(content, &result); err != nil {
			panic(new(JSONDecodeError))
        }
		return result
	}).Then(nil, func(err error) {
		if e, ok := err.(*RequestError); ok {
			fmt.Println("send request error: ", e.Error())
        } else if e, ok := err.(*JSONDecodeError); ok {
			fmt.Println("invalid json string: ", e.Error())
        } else {
			fmt.Println("error: ", err.Error())
        }
    })
}
```

#### Catch
```go
package main

import (
	"encoding/json"
	"fmt"
	"github.com/gopi-frame/future"
	"io"
	"net/http"
)

type RequestError struct{}

func (r *RequestError) Error() string {
	return "request error"
}

type JSONDecodeError struct{}

func (j *JSONDecodeError) Error() string {
	return "json decode error"
}

func main() {
	future.Async(func() map[string]any {
		resp, err := http.Get("https://example.com")
		if err != nil {
			panic(new(RequestError))
		}
		content, err := io.ReadAll(resp.Body)
		if err != nil {
			panic(err)
		}
		defer func() {
			if err := resp.Body.Close(); err != nil {
				panic(err)
			}
		}()
		var result = make(map[string]any)
		if err := json.Unmarshal(content, &result); err != nil {
			panic(new(JSONDecodeError))
        }
		return result
	}).Catch(new(RequestError), func(err error) {
		fmt.Println("send request error: ", err.Error())
	}).Catch(new(JSONDecodeError), func(err error) {
        fmt.Println("invalid json string: ", err.Error())
	}).CatchAll(func(err error) {
		fmt.Println("error: ", err.Error())
	})
}
```