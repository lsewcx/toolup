package catalog

// ossutilTool 使用阿里云 ossutil 2.0 官方二进制压缩包。
//
// 文档：https://www.alibabacloud.com/help/zh/oss/developer-reference/ossutil-overview/
// 文档中的安装步骤：下载对应平台的 zip、解压、chmod 755 ossutil、
// sudo mv ossutil /usr/local/bin/、sudo ln -s /usr/local/bin/ossutil /usr/bin/ossutil。
// 下面的下载地址与 SHA256 校验和均摘自该文档的 2.4.1 版本。
var ossutilTool = Tool{
	Name: "ossutil",
	InstallArchive: &Archive{
		Version:    "2.4.1",
		Binary:     "ossutil",
		InstallDir: "/usr/local/bin",
		Symlink:    "/usr/bin/ossutil",
		Artifacts: map[string]Artifact{
			"linux/386": {
				URL:    "https://gosspublic.alicdn.com/ossutil/v2/2.4.1/ossutil-2.4.1-linux-386.zip",
				SHA256: "b14ef9731a52c72de3c66819b451a6af494f34d73316089a17cc318c03f614e8",
			},
			"linux/amd64": {
				URL:    "https://gosspublic.alicdn.com/ossutil/v2/2.4.1/ossutil-2.4.1-linux-amd64.zip",
				SHA256: "57ee4c56796a40144acd6b0b96b53b4ba75e5cec09c01d8d0313c09a5524022f",
			},
			"linux/arm": {
				URL:    "https://gosspublic.alicdn.com/ossutil/v2/2.4.1/ossutil-2.4.1-linux-arm.zip",
				SHA256: "cc10243a4bd572716dce53ffa7d2adebe7cf74170055c83f26d72130c605ec7a",
			},
			"linux/arm64": {
				URL:    "https://gosspublic.alicdn.com/ossutil/v2/2.4.1/ossutil-2.4.1-linux-arm64.zip",
				SHA256: "be17e13f9e61ffeede0f0d3054bd0f4f146965f0f845bb8ec2ff88c5eb5638ed",
			},
			"darwin/amd64": {
				URL:    "https://gosspublic.alicdn.com/ossutil/v2/2.4.1/ossutil-2.4.1-mac-amd64.zip",
				SHA256: "455b11095c3f88ae20cedaa479a6ea03db6502fbaae9ce949d6c69c6cc7d1949",
			},
			"darwin/arm64": {
				URL:    "https://gosspublic.alicdn.com/ossutil/v2/2.4.1/ossutil-2.4.1-mac-arm64.zip",
				SHA256: "8ef710374d051817fe4483a34bf9fcdc26c1decb62dd69573f5e40d1485b45c1",
			},
		},
	},
}
