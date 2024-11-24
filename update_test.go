package betterfieldmask

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/testing/protocmp"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
)

func TestUpdate(t *testing.T) {
	tests := []struct {
		name       string
		old        proto.Message
		update     proto.Message
		updateMask *fieldmaskpb.FieldMask
		expected   proto.Message
	}{
		{
			name: "basic",
			old: &TestRoot{
				Nested: &TestRoot_Nested{
					SomeString: "old",
				},
			},
			update: &TestRoot{
				Nested: &TestRoot_Nested{
					SomeString: "new",
				},
			},
			updateMask: &fieldmaskpb.FieldMask{
				Paths: []string{"nested.some_string"},
			},
			expected: &TestRoot{
				Nested: &TestRoot_Nested{
					SomeString: "new",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Update(tt.old, tt.update, tt.updateMask)

			if diff := cmp.Diff(tt.expected, got, protocmp.Transform()); diff != "" {
				t.Fatalf("not equal (-want +got):\n%s", diff)
			}
		})
	}
}
