package main

import (
    "fmt"
    "bytes"
    "time"
    "net/http"
    "io/ioutil"
    "math/rand"
)

const desmosUrl = "https://www.desmos.com"

type desmosExpression struct {
    Type string `json:"type"`
    Id string `json:"id"`
    Color string `json:"color"`
    Latex string `json:"latex"`
}

// used for uploading the graph itself
type desmosGraph struct {
    Version int `json:"version"`
    Graph struct {
        Viewport struct {
            Xmin float64 `json:"xmin"`
            Ymin float64 `json:"ymin"`
            Xmax float64 `json:"xmax"`
            Ymax float64 `json:"ymax"`
        } `json:"viewport"`
    } `json:"graph"`
    Expressions struct {
        List []desmosExpression `json:"list"`
    } `json:"expressions"`
}

func desmosGet(url string) string {
    req, err := http.NewRequest("GET", fmt.Sprintf("%s%s", desmosUrl, url), nil)
    if err != nil {
        panic(err)
    }

    cookie := http.Cookie{Name: "sid.prod2", Value: sessionId}
    req.AddCookie(&cookie)

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        panic(err)
    }
    defer resp.Body.Close()

    body, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        panic(err)
    }

    return string(body)
}

func desmosPost(url string, data string) string {
    req, err := http.NewRequest("POST", fmt.Sprintf("%s%s", desmosUrl, url), bytes.NewBuffer([]byte(data)))
    if err != nil {
        panic(err)
    }

    cookie := http.Cookie{Name: "sid.prod2", Value: sessionId}
    req.AddCookie(&cookie)
    req.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        panic(err)
    }
    defer resp.Body.Close()

    body, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        panic(err)
    }

    return string(body)
}

func desmosRandomHash() string {
    const chars = "abcdefghijklmnopqrstuvwxyz0123456789"
    var str string
    randSource := rand.NewSource(time.Now().UnixNano())
    randGen := rand.New(randSource)
    for i := 0; i < 10; i++ {
        //strBuilder.WriteString(string(chars[randGen.Intn(len(chars))]))
        str += string(chars[randGen.Intn(len(chars))])
    }
    return str
}