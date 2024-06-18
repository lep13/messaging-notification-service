package secretsmanager

import (
	"reflect"
	"testing"

	"github.com/lep13/messaging-notification-service/models"
)

func TestGetSecretData(t *testing.T) {
	type args struct {
		secretName string
	}
	tests := []struct {
		name    string
		args    args
		want    models.SecretData
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetSecretData(tt.args.secretName)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetSecretData() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetSecretData() = %v, want %v", got, tt.want)
			}
		})
	}
}
