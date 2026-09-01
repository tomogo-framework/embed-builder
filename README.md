# embedbuilder

`embedbuilder` is a framework-agnostic Go builder for Discord rich embeds. It produces plain structs with
Discord-compatible JSON tags, so it works with any HTTP or Discord library. It requires Go 1.27 or newer.

```go
embed, err := embedbuilder.New().
	SetTitle("Deployment complete").
	SetDescription("Version 2.4.0 is live.").
	SetColor(embedbuilder.ColorBlurple).
	SetAuthor("Release bot", "https://example.com", "https://example.com/bot.png").
	AddField("Environment", "production", true).
	AddField("Status", "healthy", true).
	SetFooter("Automated notification").
	WithCurrentTimestamp().
	Build()
if err != nil {
	// A *embedbuilder.ValidationError reports every violated limit.
}

payload, err := json.Marshal(embed)
```

`Build` always calls `Validate` and returns an independent snapshot. You can also call `Validate` explicitly while
composing an embed. Validation covers text limits, required values, colors, UTF-8, timestamps, and Discord-supported URL
schemes.

The builder also supports:

- character-budget inspection with `CharacterCount`, `RemainingCharacters`, `FieldCount`, and `CanAddField`;
- indexed field replacement, insertion, and removal;
- independent import and reuse with `NewFromEmbed`, `Clone`, and `Reset`;
- RGB, hexadecimal, and official Discord brand colors; and
- validated `BuildJSON` and allocation-conscious `AppendJSON` output.

For multiple embeds, `NewCollection` provides message-level building and JSON output while enforcing Discord's ten-embed
and shared 6,000-character limits:

```go
collection := embedbuilder.NewCollection()
if err := collection.AddBuilder(firstBuilder); err != nil {
	return err
}

if err := collection.AddBuilder(secondBuilder); err != nil {
	return err
}

payload, err := collection.BuildJSON()
```

Character counts use Unicode code points and ignore leading and trailing whitespace, matching Discord's documented limit
behavior.

The limits are sourced from Discord's official
[Embed Limits](https://docs.discord.com/developers/resources/message#embed-limits) documentation.
