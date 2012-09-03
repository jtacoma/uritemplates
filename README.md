uritemplates
============

Package **uritemplates** is a level 4 implementation of [RFC 6570 (URI Template)](http://tools.ietf.org/html/rfc6570) in Go.

Installation
------------

    go get github.com/jtacoma/uritemplates

Usage
-----

Code:

    template, _ := uritemplates.Parse("https://api.github.com/repos{/user,repo}")
    values := make(map[string]interface{})
    values["user"] = "jtacoma"
    values["repo"] = "uritemplates"
    expanded, _ := template.Expand(values)
    fmt.Printf(expanded)

Output:

    https://api.github.com/repos/jtacoma/uritemplates

