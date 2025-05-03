package projects

import (
	"reflect"
	"testing"
)

func TestFilterProjectsByTopics(t *testing.T) {
	projects := []Project{
		{Name: "Project1", Topics: []string{"go", "cli"}},
		{Name: "Project2", Topics: []string{"python", "web"}},
		{Name: "Project3", Topics: []string{"go", "web"}},
		{Name: "Project4", Topics: []string{"java", "cli"}},
	}

	tests := []struct {
		name     string
		topics   []string
		expected []Project
	}{
		{
			name:     "Include Go projects",
			topics:   []string{"go"},
			expected: []Project{projects[0], projects[2]},
		},
		{
			name:     "Include CLI projects",
			topics:   []string{"cli"},
			expected: []Project{projects[0], projects[3]},
		},
		{
			name:     "Include Go and CLI projects",
			topics:   []string{"+go", "+cli"},
			expected: []Project{projects[0]},
		},
		{
			name:     "Exclude Web projects",
			topics:   []string{"-web"},
			expected: []Project{projects[0], projects[3]},
		},
		{
			name:     "Include Python projects",
			topics:   []string{"python"},
			expected: []Project{projects[1]},
		},
		{
			name:     "Include Java projects",
			topics:   []string{"java"},
			expected: []Project{projects[3]},
		},
		{
			name:     "Include all projects",
			topics:   []string{},
			expected: projects,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filtered := FilterProjectsByTopics(projects, tt.topics)
			if !reflect.DeepEqual(filtered, tt.expected) {
				t.Errorf("Expected %v, but got %v", tt.expected, filtered)
			}
		})
	}
}
