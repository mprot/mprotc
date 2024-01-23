package gen

import "testing"

func TestTitleFirstWord(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		expected string
	}{
		{
			name:     "empty string",
			text:     "",
			expected: "",
		},
		{
			name:     "single word",
			text:     "foo",
			expected: "Foo",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := TitleFirstWord(test.text)
			if got != test.expected {
				t.Errorf("unexpected text: %q (expected %q)", got, test.expected)
			}
		})
	}
}
