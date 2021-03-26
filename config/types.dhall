let Cache/Entry
    : Type
    = { on_start : Bool, every_seconds : Natural }

let Cache/Providers
    : Type
    = { update_repositories : Optional Cache/Entry
      , update_repositories_refs : Optional Cache/Entry
      }

let Cache/Slack
    : Type
    = { update_users_emails : Optional Cache/Entry }

let Cache
    : Type
    = { providers : Optional Cache/Providers, slack : Optional Cache/Slack }

let Log/Format = < json | text >

let Log/Level = < trace | debug | info | warning | error | fatal | panic >

let Log
    : Type
    = { format : Log/Format, level : Log/Level }

let Provider/Type = < github | gitlab >

let Provider
    : Type
    = { type : Provider/Type
      , url : Optional Text
      , token : Text
      , owners : List Text
      }

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
    = { cache : Optional Cache
      , log : Optional Log
      , providers : Providers
      , slack : Optional Slack
      , users : Users
      }

in  { Cache
    , Cache/Entry
    , Cache/Providers
    , Cache/Slack
    , Config
    , Log
    , Log/Format
    , Log/Level
    , Provider
    , Provider/Type
    , Providers
    , Slack
    , User
    , Users
    }
