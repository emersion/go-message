# go-message

[![GoDoc](https://godoc.org/github.com/emersion/go-message?status.svg)](https://godoc.org/github.com/emersion/go-message)
[![Build Status](https://travis-ci.org/emersion/go-message.svg?branch=master)](https://travis-ci.org/emersion/go-message)
[![codecov](https://codecov.io/gh/emersion/go-message/branch/master/graph/badge.svg)](https://codecov.io/gh/emersion/go-message)
[![Go Report Card](https://goreportcard.com/badge/github.com/emersion/go-message)](https://goreportcard.com/report/github.com/emersion/go-message)
[![Unstable](https://img.shields.io/badge/stability-unstable-yellow.svg)](https://github.com/emersion/stability-badges#unstable)

A Go library for the Internet Message Format. It implements:
* [RFC 5322](https://tools.ietf.org/html/rfc5322): Internet Message Format
* [RFC 2045](https://tools.ietf.org/html/rfc2045), [RFC 2046](https://tools.ietf.org/html/rfc2046) and [RFC 2047](https://tools.ietf.org/html/rfc2047): Multipurpose Internet Mail Extensions
* [RFC 2183](https://tools.ietf.org/html/rfc2183): Content-Disposition Header Field

## Features

* Streaming API
* Automatic encoding and charset handling
* A [`mail`](https://godoc.org/github.com/emersion/go-message/mail) subpackage to read and write mail messages

## License

MIT
