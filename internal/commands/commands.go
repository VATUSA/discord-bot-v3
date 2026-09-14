package commands

// Command is a request from the web endpoints for the bot to do something.
type Command interface {
	isCommand()
}

// SyncMember syncs the nickname and roles of a Discord user in every configured guild.
type SyncMember struct {
	UserID string
}

func (SyncMember) isCommand() {}
