package engine

type DialogueEntry struct {
	Turn int    `json:"turn"`
	From string `json:"from"`
	To   string `json:"to"`
	Text string `json:"text"`
}

const maxDialogueLog = 120

func (w *World) recordDialogue(from, to, text string) {
	w.dialogueMu.Lock()
	defer w.dialogueMu.Unlock()
	w.dialogueLog = append(w.dialogueLog, DialogueEntry{Turn: w.currentTurn, From: from, To: to, Text: text})
	if len(w.dialogueLog) > maxDialogueLog {
		w.dialogueLog = w.dialogueLog[len(w.dialogueLog)-maxDialogueLog:]
	}
}

func (w *World) dialogueSnapshot() []DialogueEntry {
	w.dialogueMu.Lock()
	defer w.dialogueMu.Unlock()
	out := make([]DialogueEntry, len(w.dialogueLog))
	copy(out, w.dialogueLog)
	return out
}
