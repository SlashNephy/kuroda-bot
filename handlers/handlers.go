package handlers

var Handlers = []any{
	onReady,
	onInteractionCreate,
	onMessageCreate,
	messageCache.OnMessageCreate,
	messageCache.OnMessageUpdate,
	postSummaryMessage.OnMessageCreate,
	postSummaryMessage.OnMessageUpdate,
	postSummaryMessage.OnMessageDelete,
}
