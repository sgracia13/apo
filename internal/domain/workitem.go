package domain

// WorkItem represents an Azure DevOps work item.
type WorkItem struct {
	ID     int                    `json:"id"`
	Rev    int                    `json:"rev"`
	Fields map[string]interface{} `json:"fields"`
	URL    string                 `json:"url"`
}

// GetField retrieves a field value as a string.
func (w *WorkItem) GetField(name string) string {
	if w.Fields == nil {
		return ""
	}
	if val, ok := w.Fields[name]; ok {
		switch v := val.(type) {
		case string:
			return v
		case map[string]interface{}:
			if displayName, ok := v["displayName"].(string); ok {
				return displayName
			}
		}
	}
	return ""
}

// Title returns the work item title.
func (w *WorkItem) Title() string {
	return w.GetField("System.Title")
}

// State returns the work item state.
func (w *WorkItem) State() string {
	return w.GetField("System.State")
}

// Type returns the work item type.
func (w *WorkItem) Type() string {
	return w.GetField("System.WorkItemType")
}

// AssignedTo returns the assigned user.
func (w *WorkItem) AssignedTo() string {
	return w.GetField("System.AssignedTo")
}

// WorkItemRef is a reference to a work item.
type WorkItemRef struct {
	ID  int    `json:"id"`
	URL string `json:"url"`
}

// WorkItemList is the response from a WIQL query.
type WorkItemList struct {
	WorkItems []WorkItemRef `json:"workItems"`
}

// WorkItemBatch is the response from batch fetching.
type WorkItemBatch struct {
	Count int        `json:"count"`
	Value []WorkItem `json:"value"`
}
