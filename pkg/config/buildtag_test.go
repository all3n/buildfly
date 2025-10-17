package config

import (
	"testing"
)

func TestParseBuildTag(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    *BuildTag
		wantErr bool
	}{
		{
			name:  "basic linux build tag",
			input: "arch=x86_64,platform=linux,runtime=glibc_2.35,compiler=gcc_11,std=cpp17,abi=sysv",
			want: &BuildTag{
				Arch:     "x86_64",
				Platform: "linux",
				Runtime:  "glibc_2.35",
				Compiler: "gcc_11",
				Std:      "cpp17",
				ABI:      "sysv",
			},
			wantErr: false,
		},
		{
			name:  "macos build tag",
			input: "arch=arm64,platform=darwin,runtime=libcxx_15,compiler=apple-clang_15.0,std=cpp20,abi=macho,target=macos14.0",
			want: &BuildTag{
				Arch:     "arm64",
				Platform: "darwin",
				Runtime:  "libcxx_15",
				Compiler: "apple-clang_15.0",
				Std:      "cpp20",
				ABI:      "macho",
				Target:   "macos14.0",
			},
			wantErr: false,
		},
		{
			name:  "cuda build tag",
			input: "arch=x86_64,platform=linux,runtime=glibc_2.35,compiler=nvcc_12.0+gcc_11.4,std=cpp17,abi=sysv,cuda=12.0,cuda_arch=compute_80|compute_90,gpu_enabled=true,gpu_backend=cuda",
			want: &BuildTag{
				Arch:     "x86_64",
				Platform: "linux",
				Runtime:  "glibc_2.35",
				Compiler: "nvcc_12.0+gcc_11.4",
				Std:      "cpp17",
				ABI:      "sysv",
				GPU: &GPUInfo{
					Backend: "cuda",
					CUDA: &CUDABackend{
						Version: "12.0",
						Arch:    []string{"compute_80", "compute_90"},
						Enabled: true,
					},
				},
			},
			wantErr: false,
		},
		{
			name:    "empty string",
			input:   "",
			want:    nil,
			wantErr: true,
		},
		{
			name:    "invalid format",
			input:   "invalid-tag",
			want:    nil,
			wantErr: true,
		},
		{
			name:    "unknown key",
			input:   "unknown=value",
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseBuildTag(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseBuildTag() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}

			if got.Arch != tt.want.Arch {
				t.Errorf("ParseBuildTag().Arch = %v, want %v", got.Arch, tt.want.Arch)
			}
			if got.Platform != tt.want.Platform {
				t.Errorf("ParseBuildTag().Platform = %v, want %v", got.Platform, tt.want.Platform)
			}
			if got.Runtime != tt.want.Runtime {
				t.Errorf("ParseBuildTag().Runtime = %v, want %v", got.Runtime, tt.want.Runtime)
			}
			if got.Compiler != tt.want.Compiler {
				t.Errorf("ParseBuildTag().Compiler = %v, want %v", got.Compiler, tt.want.Compiler)
			}
			if got.Std != tt.want.Std {
				t.Errorf("ParseBuildTag().Std = %v, want %v", got.Std, tt.want.Std)
			}
			if got.ABI != tt.want.ABI {
				t.Errorf("ParseBuildTag().ABI = %v, want %v", got.ABI, tt.want.ABI)
			}
			if got.Target != tt.want.Target {
				t.Errorf("ParseBuildTag().Target = %v, want %v", got.Target, tt.want.Target)
			}

			// 检查 GPU 配置
			if tt.want.GPU == nil && got.GPU != nil {
				t.Errorf("ParseBuildTag().GPU = %v, want nil", got.GPU)
			} else if tt.want.GPU != nil {
				if got.GPU == nil {
					t.Errorf("ParseBuildTag().GPU = nil, want %v", tt.want.GPU)
				} else {
					if got.GPU.Backend != tt.want.GPU.Backend {
						t.Errorf("ParseBuildTag().GPU.Backend = %v, want %v", got.GPU.Backend, tt.want.GPU.Backend)
					}

					if tt.want.GPU.CUDA != nil {
						if got.GPU.CUDA == nil {
							t.Errorf("ParseBuildTag().GPU.CUDA = nil, want %v", tt.want.GPU.CUDA)
						} else {
							if got.GPU.CUDA.Version != tt.want.GPU.CUDA.Version {
								t.Errorf("ParseBuildTag().GPU.CUDA.Version = %v, want %v", got.GPU.CUDA.Version, tt.want.GPU.CUDA.Version)
							}
							if !equalStringSlices(got.GPU.CUDA.Arch, tt.want.GPU.CUDA.Arch) {
								t.Errorf("ParseBuildTag().GPU.CUDA.Arch = %v, want %v", got.GPU.CUDA.Arch, tt.want.GPU.CUDA.Arch)
							}
							if got.GPU.CUDA.Enabled != tt.want.GPU.CUDA.Enabled {
								t.Errorf("ParseBuildTag().GPU.CUDA.Enabled = %v, want %v", got.GPU.CUDA.Enabled, tt.want.GPU.CUDA.Enabled)
							}
						}
					}
				}
			}
		})
	}
}

func TestBuildTagString(t *testing.T) {
	tests := []struct {
		name string
		bt   *BuildTag
		want string
	}{
		{
			name: "basic linux",
			bt: &BuildTag{
				Arch:     "x86_64",
				Platform: "linux",
				Runtime:  "glibc_2.35",
				Compiler: "gcc_11",
				Std:      "cpp17",
				ABI:      "sysv",
			},
			want: "arch=x86_64,platform=linux,runtime=glibc_2.35,compiler=gcc_11,std=cpp17,abi=sysv",
		},
		{
			name: "with cuda",
			bt: &BuildTag{
				Arch:     "x86_64",
				Platform: "linux",
				Runtime:  "glibc_2.35",
				Compiler: "nvcc_12.0+gcc_11.4",
				Std:      "cpp17",
				ABI:      "sysv",
				GPU: &GPUInfo{
					Backend: "cuda",
					CUDA: &CUDABackend{
						Version: "12.0",
						Arch:    []string{"compute_80", "compute_90"},
						Enabled: true,
					},
				},
			},
			want: "arch=x86_64,platform=linux,runtime=glibc_2.35,compiler=nvcc_12.0+gcc_11.4,std=cpp17,abi=sysv,cuda=12.0,cuda_arch=compute_80|compute_90",
		},
		{
			name: "nil build tag",
			bt:   nil,
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.bt.String()
			if got != tt.want {
				t.Errorf("BuildTag.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBuildTagToDirName(t *testing.T) {
	tests := []struct {
		name string
		bt   *BuildTag
		want string
	}{
		{
			name: "basic linux",
			bt: &BuildTag{
				Arch:     "x86_64",
				Platform: "linux",
				Runtime:  "glibc_2.35",
				Compiler: "gcc_11",
				Std:      "cpp17",
				ABI:      "sysv",
			},
			want: "arch-x86_64,platform-linux,runtime-glibc_2.35,compiler-gcc_11,std-cpp17,abi-sysv",
		},
		{
			name: "with plus and pipe",
			bt: &BuildTag{
				Arch:     "x86_64",
				Platform: "linux",
				Runtime:  "glibc_2.35+",
				Compiler: "gcc_11+",
				Std:      "cpp17",
				ABI:      "sysv",
				GPU: &GPUInfo{
					Backend: "cuda",
					CUDA: &CUDABackend{
						Version: "12.0",
						Arch:    []string{"compute_80", "compute_90"},
						Enabled: true,
					},
				},
			},
			want: "arch-x86_64,platform-linux,runtime-glibc_2.35plus,compiler-gcc_11plus,std-cpp17,abi-sysv,cuda-12.0,cuda_arch-compute_80orcompute_90",
		},
		{
			name: "nil build tag",
			bt:   nil,
			want: "default",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.bt.ToDirName()
			if got != tt.want {
				t.Errorf("BuildTag.ToDirName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBuildTagValidate(t *testing.T) {
	tests := []struct {
		name    string
		bt      *BuildTag
		wantErr bool
	}{
		{
			name: "valid linux build tag",
			bt: &BuildTag{
				Arch:     "x86_64",
				Platform: "linux",
				Runtime:  "glibc_2.35",
				Compiler: "gcc_11",
				Std:      "cpp17",
				ABI:      "sysv",
			},
			wantErr: false,
		},
		{
			name: "valid macos build tag",
			bt: &BuildTag{
				Arch:     "arm64",
				Platform: "darwin",
				Runtime:  "libcxx_15",
				Compiler: "apple-clang_15.0",
				Std:      "cpp20",
				ABI:      "macho",
			},
			wantErr: false,
		},
		{
			name: "invalid arch",
			bt: &BuildTag{
				Arch:     "invalid_arch",
				Platform: "linux",
				Runtime:  "glibc_2.35",
				Compiler: "gcc_11",
				Std:      "cpp17",
				ABI:      "sysv",
			},
			wantErr: true,
		},
		{
			name: "invalid platform",
			bt: &BuildTag{
				Arch:     "x86_64",
				Platform: "invalid_platform",
				Runtime:  "glibc_2.35",
				Compiler: "gcc_11",
				Std:      "cpp17",
				ABI:      "sysv",
			},
			wantErr: true,
		},
		{
			name: "invalid std",
			bt: &BuildTag{
				Arch:     "x86_64",
				Platform: "linux",
				Runtime:  "glibc_2.35",
				Compiler: "gcc_11",
				Std:      "invalid_std",
				ABI:      "sysv",
			},
			wantErr: true,
		},
		{
			name:    "nil build tag",
			bt:      nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.bt.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("BuildTag.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestBuildTagEquals(t *testing.T) {
	bt1 := &BuildTag{
		Arch:     "x86_64",
		Platform: "linux",
		Runtime:  "glibc_2.35",
		Compiler: "gcc_11",
		Std:      "cpp17",
		ABI:      "sysv",
	}

	bt2 := &BuildTag{
		Arch:     "x86_64",
		Platform: "linux",
		Runtime:  "glibc_2.35",
		Compiler: "gcc_11",
		Std:      "cpp17",
		ABI:      "sysv",
	}

	bt3 := &BuildTag{
		Arch:     "arm64",
		Platform: "linux",
		Runtime:  "glibc_2.35",
		Compiler: "gcc_11",
		Std:      "cpp17",
		ABI:      "sysv",
	}

	tests := []struct {
		name string
		bt1  *BuildTag
		bt2  *BuildTag
		want bool
	}{
		{
			name: "equal build tags",
			bt1:  bt1,
			bt2:  bt2,
			want: true,
		},
		{
			name: "different build tags",
			bt1:  bt1,
			bt2:  bt3,
			want: false,
		},
		{
			name: "both nil",
			bt1:  nil,
			bt2:  nil,
			want: true,
		},
		{
			name: "one nil",
			bt1:  bt1,
			bt2:  nil,
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.bt1.Equals(tt.bt2)
			if got != tt.want {
				t.Errorf("BuildTag.Equals() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsValidBuildTagString(t *testing.T) {
	tests := []struct {
		name string
		s    string
		want bool
	}{
		{
			name: "valid basic tag",
			s:    "arch=x86_64,platform=linux",
			want: true,
		},
		{
			name: "valid complex tag",
			s:    "arch=x86_64,platform=linux,runtime=glibc_2.35,compiler=gcc_11,std=cpp17,abi=sysv",
			want: true,
		},
		{
			name: "empty string",
			s:    "",
			want: false,
		},
		{
			name: "invalid format - no equals",
			s:    "invalid-tag",
			want: false,
		},
		{
			name: "invalid format - starts with number",
			s:    "1arch=x86_64",
			want: false,
		},
		{
			name: "invalid format - empty key",
			s:    "=value",
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsValidBuildTagString(tt.s)
			if got != tt.want {
				t.Errorf("IsValidBuildTagString() = %v, want %v", got, tt.want)
			}
		})
	}
}

// 辅助函数
func equalStringSlices(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
