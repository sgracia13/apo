package domain

// Pipeline represents an Azure DevOps pipeline.
type Pipeline struct {
	ID     int    `json:"id"`
	Name   string `json:"name"`
	Folder string `json:"folder"`
	URL    string `json:"url"`
}

// FullPath returns the complete path including folder.
func (p *Pipeline) FullPath() string {
	if p.Folder == "" || p.Folder == "\\" || p.Folder == "/" {
		return p.Name
	}
	return p.Folder + "/" + p.Name
}

// PipelineList is the response from listing pipelines.
type PipelineList struct {
	Count int        `json:"count"`
	Value []Pipeline `json:"value"`
}
