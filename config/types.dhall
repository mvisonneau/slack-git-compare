let Log/Format = < json | text >

let Log/Level = < trace | debug | info | warning | error | fatal | panic >

let Log
    : Type
    = { format : Log/Format, level : Log/Level }

let Provider/Type = < github | gitlab >

let Provider
    : Type
    = { type: Provider/Type, url : Optional Text, token : Text, owners : List Text }

let Providers
    : Type
    = List Provider

let Slack
    : Type
    = { token : Text, signing_secret : Text }

let User
    : Type
    = { email : Text, aliases : List Text }

let Users
    : Type
    = List User

let Config
    : Type
    = { providers : Providers
      , slack : Optional Slack
      , log : Optional Log
      , users : Users
      }

in

{ Log/Format
, Log/Level
, Log
, Provider
, Providers
, Provider/Type
, Slack
, User
, Users
, Config
}