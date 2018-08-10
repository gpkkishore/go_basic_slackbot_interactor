This is simple GO program for listening messages from slack bot and parsing with simple regex
<bot> help


File Structure

sre_slack.go - Contains Go code interacting with bot user
files - Contains required files for GO program to run
files/bot_token.json - Bot token has to be mentioned
channels_whitelist - For explicitly whitelisting channels to run/listen the bot
users_whitelist - For explicitly whitelisting users to run bot commands