let T = ./types.dhall

let cfg
    : T.Config
    = { cache = Some
        { providers = Some
          { update_repositories = Some { on_start = True, every_seconds = 3600 }
          , update_repositories_refs = Some
            { on_start = False, every_seconds = 0 }
          }
        , slack = Some
          { update_users_emails = Some { on_start = True, every_seconds = 86400 }
          }
        }
      , log = Some { format = T.Log/Format.text, level = T.Log/Level.debug }
      , providers =
        [ { type = T.Provider/Type.github
          , url = None Text
          , token = "xxxx"
          , owners = [ "cilium" ]
          }
        , { type = T.Provider/Type.gitlab
          , url = None Text
          , token = "xxxx"
          , owners = [ "gitlab-org" ]
          }
        ]
      , slack = Some { token = "xobt-xxxxxx", signing_secret = "xxxxx" }
      , users =
        [ { email = "foo@bar.baz"
          , aliases = [ "alice@yolo.com", "bob@yolo.com" ]
          }
        ]
      }

in  cfg
