# Karma

Karma is a webhook server for adding "karma points" to a set of discord servers.
When registered with your Discord server, it will add the following commands:

`gib` - Awards one point to another user in the server accompanied by a message

`checkkarma` - Checks a given users current total of karma

`topten` - Checks the most awarded users' karma totals

## Set up

After creating a Disord app and awarding the proper permissions (stuff relating
to responding to messages), point the `Interactions Endpoint Url` to your
server.

Here's are the env vars your server will need set:

| Name | Value |
----------------
| `PORT` | What port the server should listen on |
| `TLS_CERT_FILE` | Cert for TLS. If you're going to put the server _behind_ an
HTTPS connection then you can omit this and it will server just HTTP. But
discord does require that the endpoint be over HTTPS |
| `TLS_KEY_FILE` | The private key file component of serving over TLS. Optional
if you're not doing that |
| `DB_PATH` | Path to the sqlite DB file |
| `DISCORD_TOKEN` | The token given by discord and used in the authorization of
calls to discord |
| `DISCORD_APP_ID` | The app id when you register the application |
| `DISCORD_GUILD_IDS` | A comma-separated list of guild ids that the server
should server for |
| `DISCORD_VERIFY_KEY` | Discord gives you a public key that you have to use to
verify their signed calls. They will send invalid requests to make sure you're
verifying calls to your server |
| `SKIP_REGISTER` | Optional. At startup, the server will call to register
commands with the given guild ID's. This can be rate limited, so if you want to
skip that, just set this to true |
