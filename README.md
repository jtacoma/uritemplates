uritemplates
============

This is an implementation-in-progress of [URI Templates (RFC 6570)](http://tools.ietf.org/html/rfc6570) in Go.

Installation
------------

    go get github.com/jtacoma/uritemplates

Usage
-----

Code:

    template, _ := uritemplates.Parse("https://api.github.com/repos/{user}/{repo}")
    values := make(map[string]interface{})
    values["user"] = "jtacoma"
    values["repo"] = "uritemplates"
    fmt.Printf(template.ExpandString(values))

Output:

    https://api.github.com/repos/jtacoma/uritemplates

