# dblens

An HTTP handler for database browsing.

BEWARE: `dblens.Handler` accepts ALL queries, including destructive ones. Don't expose the handler to the wide public.

## Usage

```shell
go get -u github.com/maragudk/dblens
```

```go
package main

import (
	"database/sql"
	"errors"
	"log"
	"net/http"
	"os"

	_ "github.com/mattn/go-sqlite3"

	"github.com/maragudk/dblens"
)

func main() {
	os.Exit(start())
}

func start() int {
	log := log.New(os.Stderr, "", log.Ldate|log.Ltime|log.Lshortfile|log.LUTC)

	db, err := sql.Open("sqlite3", "app.db?_journal=WAL&_timeout=5000&_fk=true")
	if err != nil {
		log.Println("Error opening database:", err)
		return 1
	}

	log.Println("Starting on http://localhost:8080")

	if err := http.ListenAndServe("localhost:8080", dblens.Handler(db, "sqlite3")); err != nil &&
		!errors.Is(err, http.ErrServerClosed) {
		log.Println("Error:", err)
		return 1
	}

	return 0
}
```

## Screenshot

![screenshot.png](docs%2Fscreenshot.png)

Made in 🇩🇰 by [maragu](https://www.maragu.dk/), maker of [online Go courses](https://www.golang.dk/).
