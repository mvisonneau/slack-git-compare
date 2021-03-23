let T = ./types.dhall

let cfg
    : T.Config
    = { providers =
        [ { type = T.Provider/Type.github, url = None Text, token = "xxxx", owners = [ "cilium" ] }
        , { type = T.Provider/Type.gitlab, url = None Text, token = "xxxx", owners = [ "gitlab-org" ] }
        ]
      , slack = Some { token = "xobt-xxxxxx", signing_secret = "xxxxx" }
      , log = None T.Log
      , users =
        [ { email = "foo@bar.baz"
          , aliases = [ "alice@yolo.com", "bob@yolo.com" ]
          }
        ]
      }

in  cfg