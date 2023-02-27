// Package dblens provides a Handler which returns an HTML page (or page fragment) to query a database.
// There are no authentication and authorization mechanisms included, and queries are passed verbatim to the database,
// so use with care!
package dblens

import (
	"context"
	"database/sql"
	"net/http"
	"net/url"

	"github.com/jmoiron/sqlx"
	g "github.com/maragudk/gomponents"
	"github.com/maragudk/gomponents-heroicons/v2/solid"
	hx "github.com/maragudk/gomponents-htmx"
	hxhttp "github.com/maragudk/gomponents-htmx/http"
	c "github.com/maragudk/gomponents/components"
	. "github.com/maragudk/gomponents/html"

	ghttp "github.com/maragudk/gomponents/http"
)

type table struct {
	Columns []string
	Data    [][]any
}

// Handler returns a http.Handler ready to be added to your routes.
// It returns an HTML page with a query input box and a table of results from the query, if any.
// If
func Handler(db *sql.DB, driverName string) http.HandlerFunc {
	return ghttp.Adapt(func(w http.ResponseWriter, r *http.Request) (g.Node, error) {
		query := r.URL.Query().Get("query")

		var err error
		var t table
		if query != "" {
			t, err = runQuery(r.Context(), sqlx.NewDb(db, driverName), query)
		}

		if hxhttp.IsRequest(r.Header) {
			hxhttp.SetPushURL(w.Header(), r.URL.Path+"?query="+url.QueryEscape(query))
			return result(t, err), nil
		}

		return page(r.URL.Path, query, t, err), nil
	})
}

func runQuery(ctx context.Context, db *sqlx.DB, query string) (t table, err error) {
	var rows *sqlx.Rows
	rows, err = db.QueryxContext(ctx, query)
	if err != nil {
		return
	}
	defer func() {
		if closeErr := rows.Close(); err == nil && closeErr != nil {
			err = closeErr
		}
	}()

	t.Columns, err = rows.Columns()
	if err != nil {
		return
	}

	for rows.Next() {
		var row []any
		row, err = rows.SliceScan()
		if err != nil {
			return
		}
		t.Data = append(t.Data, row)
	}

	err = rows.Err()

	return
}

func page(path, query string, t table, err error) g.Node {
	return c.HTML5(c.HTML5Props{
		Title:    "dblens",
		Language: "en",
		Head: []g.Node{
			Script(Src("https://cdn.tailwindcss.com?plugins=forms,typography")),
			Script(Src("https://unpkg.com/htmx.org")),
		},
		Body: []g.Node{
			container(
				FormEl(Action(path), Method("get"), Class("flex items-center w-full mt-16 mb-32"),
					hx.Boost("true"), hx.Target("#result"),
					hx.Swap("innerHTML show:window:top"),
					Label(For("query"), Class("sr-only"), g.Text("Query")),
					Div(Class("relative rounded-md shadow-sm flex-grow"),
						Div(Class("absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none"),
							solid.CircleStack(Class("h-5 w-5 text-gray-400")),
						),
						Input(Type("text"), Name("query"), ID("query"), Value(query), TabIndex("1"),
							AutoComplete("off"), AutoFocus(),
							Class("focus:ring-gray-500 focus:border-gray-500 block w-full pl-10 text-sm border-gray-300 rounded-md"),
						),
					),
				),

				Div(ID("result"),
					result(t, err),
				),
			),
		},
	})
}

func result(t table, err error) g.Node {
	return Div(
		g.If(err != nil, Strong(g.Textf("Error: %v", err))),

		g.If(err == nil && len(t.Columns) > 0,
			Div(Class("flex flex-col"),
				Div(Class("-my-2 -mx-4 overflow-x-auto sm:-mx-6 lg:-mx-8"),
					Div(Class("inline-block min-w-full py-2 align-middle md:px-6 lg:px-8"),
						Div(Class("overflow-hidden shadow ring-1 ring-black ring-opacity-5 md:rounded-lg"),
							Table(Class("min-w-full divide-y divide-gray-300"),
								THead(Class("bg-gray-50"),
									Tr(
										g.Map(t.Columns, func(c string) g.Node {
											return Th(
												g.Attr("scope", "col"),
												Class("px-3 py-3.5 text-left text-sm font-semibold text-gray-900"),
												g.Text(c),
											)
										})...,
									),
								),

								TBody(Class("bg-white divide-y divide-gray-200"),
									g.Group(g.Map(t.Data, func(row []any) g.Node {
										return Tr(
											g.Map(row, func(d any) g.Node {
												return Td(Class("whitespace-nowrap px-3 py-4 text-sm text-gray-900"), g.Textf("%v", d))
											})...,
										)
									})),
								),
							),
						),
					),
				),
			),
		),
	)
}

func container(children ...g.Node) g.Node {
	return Div(Class("max-w-7xl mx-auto p-4 sm:p-6 lg:p-8"), g.Group(children))
}
