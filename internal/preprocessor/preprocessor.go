package preprocessor

import (
	"github.com/geocine/geopub/internal/models"
)

// Preprocessor interface for processing chapters before rendering
type Preprocessor interface {
	Name() string
	Process(book *models.Book) error
}

// Pipeline runs multiple preprocessors in sequence
type Pipeline struct {
	preprocessors []Preprocessor
}

// NewPipeline creates a new preprocessor pipeline
func NewPipeline() *Pipeline {
	return &Pipeline{
		preprocessors: make([]Preprocessor, 0),
	}
}

// Add adds a preprocessor to the pipeline
func (p *Pipeline) Add(preprocessor Preprocessor) {
	p.preprocessors = append(p.preprocessors, preprocessor)
}

// Process runs all preprocessors on the book
func (p *Pipeline) Process(book *models.Book) error {
	for _, preprocessor := range p.preprocessors {
		if err := preprocessor.Process(book); err != nil {
			return err
		}
	}
	return nil
}
