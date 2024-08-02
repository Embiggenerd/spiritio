# SpiritIO

This is a zoom clone application that utilizes webRTC for video/audio, and
websockets for messaging and signaling. Live at
[chat.igoratakhanov.com](https://chat.igoratakhanov.com).

## Instructions

Clone repo

```
$cd go
$make docker
```

Open localhost:8080 in chrome. Open a chrome incognito instance, copy paste the
same URL with the auto-generated room ID. Talk to yourself, send messages, or
send private messages. Type '/' for a list of commands, type '@' for a list of
users in the room you can direct message.

## Features:

All instructions happen through text.

```
/login <name> <password>
/set password <password>
/set name <name>
@<username> <direct message>
```

You automatically get a temporary account with a random name. Set a password,
and then you can change the name and later login using your credentials.

If video suddenly stops working when running locally, it could be because chrome
turns off video if it thinks it's misbehaving, so restart chrome.

## Implementation

This uses no framework on the frontend - everything is vanilla JS to demonstrate
you don't need reactive state or a virtual DOM. The types are defined in a .d.ts
file, however they are referenced using JS Docs. Types but no build step or
bundle file.

The backend is a Selective Forwarding Unit, which means there are no peer to
peer connections. The browser doesn't know this, and it goes through the various
stages of RTC connection and data transfer with the backend, which implements an
RTC client via [pion](https://github.com/pion)

## Philosophy

All communication, except for video/audio streaming, is done via websockets.
There is no request/response scheme, everything is async.

The client sends "work orders" to the backend with an **order** and **details**.
The backend will do one of two things: send an event or a question. The event
message has an **event** and **data** fields. A question has an **ask** field.
There is no 1:1 relationship between messages send backand forth. The client
sends arbitrary orders, which may or may not need clarification to result in an
event.

```
WorkOrder = { order: string, details: any } // From client to backend

Event = { event: string, details: any } // From backend to client to notify of application state changes

Question = { ask: string } // When backend needs clarification
```

A lot of failover mechanisms are used to ensure a smooth product experience. For
instance, if you put an invalid roomID in the url, you will simply be redirected
to a new room.

It's very hard not to use the app, you get passwordless access with a very
simple option to register via text input, doing away with modals or other pages,
or buttons to click.

## TODO

Write tests.

Add direct messaging feature. ✓

Add jsdoc types to front end. ✓

Refactor backend for testability.

Implement better logging.
