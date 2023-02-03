# Dgbridge

dgbridge is a universal process wrapper that can bridge a process' output and input
streams to a Discord channel based on user-defined rules. The process' output
is relayed to Discord, but first it goes through a set of regex matchers and
then replaced into a template, called a **list of matching rules**.

It is  intended for the process wrapper to function with any process; be nearly
indistinguishable from the underlying server, and retain all server functionality
as if it were running without a wrapper.

### Matching Rules
#### From Process to Discord
For example, the process might output:

    12:20:50 [Worker Thread 1/INFO]: <Player> Hello!

And you might have a **matching rule** like so:

    "SubprocessToDiscord": [
        {
          "Match": ".*\\[.*INFO]:? <(.+)> (.+)",
          "Template": "**<${1}>** ${2}"
        }
    ]

If the line matches the `Match` regex, then the regex matching groups are
placed into the `Template`, and the result of the replacement is sent to the
Discord channel. In this example, the result sent to Discord would be:

    **<Player>** Hello!

#### From Discord to the Process

For example, someone in Discord might say:

    Hello!

And you might have a **matching rule** like so:

    "DiscordToSubprocess": [
        {
            "Match": ".*",
            "Template": "say <^U> $0"
        }
    ],

If the Discord message matches the `Match` regex, then the regex matching groups
are placed into the `Template`, and the result of the replacement written to the
subprocess' input stream. This example is for a Minecraft server. The template
creates a `say` command that will send the Discord message to the server through
the server console. If you need fancier formatting, you can replace `say` with 
`tellraw`.

However, do note the `^U` **Discord replacement 
parameter** in the template. **Discord -> Subprocess** templates allow for
special parameters to insert Discord-specific variables into templates.
The parameters available are:

- `^U`: Discord username of sender
- `^T`: Discord discriminator of sender (the #0000 tag)
- `^^`: Escape sequence for `^`

The bridge will replace these parameters with variables from the context of the
Discord message.

<hr>

The program comes with pre-made rules for Minecraft and Terraria servers, so
you can look at them for some examples.

## Supported Operating Systems

### Officially Supported:

- Linux
- macOS

### Untested but supported:

- FreeBSD
- NetBSD
- OpenBSD

### Unsupported:

- Windows. Unfortunately, a lot of things won't work properly on Windows because
  Windows does not support process signals, and sometimes it just acts weird for
  some reason.

## Usage

    dgbridge --token <YOUR_DISCORD_TOKEN> \
             --channel_id <CHANNEL_ID> \
             --rules <RULES_FILE> \
             <COMMAND>

## Examples

### Minecraft Example

    dgbridge --token TOKEN \
             --channel_id CHANNEL_ID \
             --rules ./rules/minecraft.rules.json \
             "java -Xms512M -Xmx1G -jar server.jar nogui"

### Terraria Example

It is recommended to use a configuration file for your Terraria server to skip
the interactive setup. See https://terraria.fandom.com/wiki/Guide:Setting_up_a_Terraria_server#Making_a_configuration_file
for how to make and use a configuration file.

    dgbridge --token TOKEN \
             --channel_id CHANNEL_ID \
             --rules ./rules/minecraft.rules.json \
             "./TerrariaServer -config config.ini"

## License

See [LICENSE](./LICENSE.txt).