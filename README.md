<!-- omit from toc -->
# Table of Contents

- [Compatible Game Servers](#compatible-game-servers)
  - [Confirmed Supported:](#confirmed-supported)
  - [Confirmed Unsupported:](#confirmed-unsupported)
  - [Supported Operating Systems](#supported-operating-systems)
    - [Officially Supported:](#officially-supported)
    - [Untested but supported:](#untested-but-supported)
    - [Unsupported:](#unsupported)
- [What is dgbridge?](#what-is-dgbridge)
- [Basic Usage](#basic-usage)
- [Examples](#examples)
  - [Minecraft Example](#minecraft-example)
  - [Terraria Example](#terraria-example)
- [Rules](#rules)
  - [Rules Example: Process ➡️ Discord](#rules-example-process-️-discord)
  - [Rules Example: Discord ➡️ Process](#rules-example-discord-️-process)
- [Automated Rule Testing](#automated-rule-testing)
- [Questions](#questions)
  - [1. How does this differ from a Discord bridge like DiscordSRV?](#1-how-does-this-differ-from-a-discord-bridge-like-discordsrv)
  - [2. Is this supported on the platform I'm using (e.g.: Pterodactyl Panel)?](#2-is-this-supported-on-the-platform-im-using-eg-pterodactyl-panel)
  - [3. Does this really work with any game server?](#3-does-this-really-work-with-any-game-server)
  - [4. How do I make my own rules?](#4-how-do-i-make-my-own-rules)
- [Contributing](#contributing)
- [License](#license)
  
<!-- omit from toc -->
# Dgbridge
<!-- omit from toc -->
## A universal Discord bridge for Minecraft and more!

  - Written in Golang
  - Open-source
  - Tiny source code
  - MIT-licensed
  - Automated testing tool included

# Compatible Game Servers

  **If there are any game servers you have tested that is not listed here, please let me know how your setup went so that we can expand this list.**

## Confirmed Supported:

  - Minecraft Alpha 1.1.2_01
  - Minecraft Beta 1.3_01
  - Minecraft Release 1.19.3
  - Minecraft Forge 1.19
  - Terraria

## Confirmed Unsupported:

  - Minecraft Bedrock
    - Reason: Server does not print player messages to console and does not provide such feature

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

# What is dgbridge?

dgbridge is a universal process wrapper that can bridge a process' output and input
streams to a Discord channel based on user-defined rules. The process' output
is relayed to Discord, but first it goes through a set of regex matchers and
then replaced into a template, called a **list of matching rules**.

It is intended for the process wrapper to function with any process; be nearly
indistinguishable from the underlying server, and retain all server functionality
as if it were running without a wrapper.

# Basic Usage

    dgbridge --token <YOUR_DISCORD_TOKEN> \
             --channel_id <CHANNEL_ID> \
             --rules <RULES_FILE> \
             <COMMAND>

# Examples

## Minecraft Example

    dgbridge --token TOKEN \
             --channel_id CHANNEL_ID \
             --rules ./rules/minecraft.rules.json \
             "java -Xms512M -Xmx1G -jar server.jar nogui"

## Terraria Example

It is recommended to use a configuration file for your Terraria server to skip
the interactive setup. See https://terraria.fandom.com/wiki/Guide:Setting_up_a_Terraria_server#Making_a_configuration_file
for how to make and use a configuration file.

    dgbridge --token TOKEN \
             --channel_id CHANNEL_ID \
             --rules ./rules/terraria.rules.json \
             "./TerrariaServer -config config.ini"

# Rules

Rules tell dgbridge how to translate process output to Discord output and vice-versa.
You may see some basic rules in the [rules/](./rules/) directory.
The rules included should cover the most basic needs, but some advanced users may need to tinker with their rules.

## Rules Example: Process ➡️ Discord

This is an example of how a basic **Process ➡️ Discord** rule works.

Given the console output:

    12:20:50 [Worker Thread 1/INFO]: <Player> Hello!

And the rule:

    "SubprocessToDiscord": [
        {
          "Match": ".*\\[.*INFO]:? <(.+)> (.+)",
          "Template": "**<${1}>** ${2}"
        }
    ]

If the console output matches the `Match` regex, then the regex matching groups (`${1}` and `${2}` in this example) are replaced in the `Template`, and the result of the replacement is sent to the Discord channel. In this example, the result sent to Discord would be:

    **<Player>** Hello!

## Rules Example: Discord ➡️ Process

This is an example of how a basic **Discord ➡️ Process** rule works.
Take for example, a Discord message containing the text:

    Hello!

And the rule:

    "DiscordToSubprocess": [
        {
            "Match": ".*",
            "Template": "say <^U> $0"
        }
    ],

If the Discord message matches the `Match` regex, then the regex matching groups
are placed into the `Template`, and the result of the replacement written to the
subprocess' input stream. This particular rule will use the `say` command to send
a message to a Minecraft server.

However, do note the `^U` **parameter** in the template. **Discord ➡️ Subprocess** templates allow for special parameters to insert Discord-specific variables into templates.
The parameters available are:

- `^U`: Discord username of sender
- `^T`: Discord discriminator of sender (the #0000 tag)
- `^C`: Discord user's display color
- `^^`: Escape sequence for `^`

The bridge will replace these parameters with variables from the context of the
Discord message.

<hr>

The program comes with pre-made rules for Minecraft and Terraria servers, so
you can look at them for some more examples.

# Automated Rule Testing

You can automate rule testing with the Dgbridge Testing Tool. The ruletester program accepts a rules file and a test case file. It checks that all regex rules are applied in the way you expect them to.

Example usage:

```
./ruletester --rules ../rules/minecraft.rules.json --test ../tests/test.minecraft.rules.json
Dgbridge Rule Tester (v1.0.1)
-------------------------------------------
SubprocessToDiscord tests: Running 8 tests
-------------------------------------------
✅  Test #0: PASS
✅  Test #1: PASS
✅  Test #2: PASS
✅  Test #3: PASS
❌  SubprocessToDiscordTest Test #4: FAIL:
        Input:          [26Apr2023 05:52:38.452] [Server thread/INFO] [net.minecraft.server.dedicated.DedicatedServer/]: Bob left the game
        Expected:       :arrow_left: **Bob** left.
        Got:            :arrow_left: **Bob** disconnected.
✅  Test #5: PASS
✅  Test #6: PASS
✅  Test #7: PASS
-------------------------------------------
DiscordToSubprocess tests: Running 2 tests
-------------------------------------------
✅  Test #0: PASS
✅  Test #1: PASS
Finished: Tests passed: 9, failed: 1
```

See the `tests/test.minecraft.rules.json` for an example of a test case.

# Questions

## 1. How does this differ from a Discord bridge like DiscordSRV?

A: A traditional Discord bridge for Minecraft usually comes in the form of a server plugin, or a Forge mod, or a Fabric mod. These are all Minecraft server modifications that interact with server code. Dgbridge is none of those, and I’ll get to that in a second.

DiscordSRV is a Spigot plugin. A plugin interacts with the Minecraft server by injecting its code into the server binary. But the server code is different for every version, so the plugin developer has to recompile and potentially make major code changes to support a new version. There is little interest for plugin developers to support older Minecraft server versions, so it is likely that you’ll never see a version DiscordSRV that’s compatible with Alpha, or Beta, or some older version of release. It’s also likely that you’ll have to wait for the developer to update their plugin after a new update that renders the plugin incompatible. It’s also only compatible with Spigot, so if you were using some other server, such as Forge, you’d need to find an alternative compatible with Forge.

These are the problems that dgbridge aims to solve. By not being dependent on the server code, it doesn’t matter what version of the server what you’re using. It makes no difference whether you’re using Spigot, Paper, Forge, Fabric — dgbridge works on all of them by simply wrapping the process and using the console to transmit data.

## 2. Is this supported on the platform I'm using (e.g.: Pterodactyl Panel)?

It should run on any platform where you're allowed to specify the server executable to be used. It's in my best interest for this to be accessible, so hit me up with questions if you have any.

## 3. Does this really work with any game server?

Maybe. It should work with anything that outputs chat messages to the console and also lets the server send messages through the console.

## 4. How do I make my own rules?

You will need to be familiar with Regex (regular expression language) to understand and use the rules to their full potential.

You may find this helpful: 

 - **Regular Expressions from MDN Web Docs** https://developer.mozilla.org/en-US/docs/Web/JavaScript/Guide/Regular_expressions
 - **RegexOne: Learn Regex with interactive exercises** https://regexone.com/

# Contributing

I appreciate any feedback, feature requests and pull requests. Please use the Issues tab for discussion.

# License

See [LICENSE](./LICENSE.txt).