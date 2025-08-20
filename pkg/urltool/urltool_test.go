package urltool

import "testing"

func TestGetBasePath(t *testing.T) {
	type args struct {
		targeturl string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{name: "正确示例", args: args{targeturl: "https://www.liwenzhou.com/posts/Go/golang-menu/"}, want: "golang-menu", wantErr: false},
		{name: "无效的url示例", args: args{targeturl: "/xxxx/1123"}, want: "", wantErr: true},
		{name: "空字符串", args: args{targeturl: ""}, want: "", wantErr: true},
		{name: "带query的url", args: args{targeturl: "https://www.liwenzhou.com/posts/Go/golang-menu/?a=1&b=2"}, want: "golang-menu", wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetBasePath(tt.args.targeturl)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetBasePath() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("GetBasePath() got = %v, want %v", got, tt.want)
			}
		})
	}
}
