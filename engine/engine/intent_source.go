package engine

type IntentSource string

const (
	SourceAI IntentSource = "ai"

	SourceAIBatch IntentSource = "ai-batch"

	SourceAgentBatch IntentSource = "agent_batch"

	SourceInstinct IntentSource = "instinct"

	SourceRoutine IntentSource = "routine"

	SourceScripted IntentSource = "scripted"

	SourceExternal IntentSource = "external"

	SourcePlayerIdle IntentSource = "player_idle"

	SourceFallbackBatch        IntentSource = "fallback_batch"
	SourceFallbackBatchMissing IntentSource = "fallback_batch_missing"
)
