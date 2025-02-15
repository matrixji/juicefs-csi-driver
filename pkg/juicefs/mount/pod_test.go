/*
Copyright 2021 Juicedata Inc

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package mount

import (
	"reflect"
	"testing"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/juicedata/juicefs-csi-driver/pkg/juicefs/config"
)

var (
	podLimit = map[corev1.ResourceName]resource.Quantity{
		corev1.ResourceCPU:    resource.MustParse("1"),
		corev1.ResourceMemory: resource.MustParse("2G"),
	}
	podRequest = map[corev1.ResourceName]resource.Quantity{
		corev1.ResourceCPU:    resource.MustParse("3"),
		corev1.ResourceMemory: resource.MustParse("4G"),
	}
	testResources = corev1.ResourceRequirements{
		Limits:   podLimit,
		Requests: podRequest,
	}
)

func Test_parsePodResources(t *testing.T) {
	type args struct {
		MountPodCpuLimit   string
		MountPodMemLimit   string
		MountPodCpuRequest string
		MountPodMemRequest string
	}
	tests := []struct {
		name string
		args args
		want corev1.ResourceRequirements
	}{
		{
			name: "test",
			args: args{
				MountPodCpuLimit:   "1",
				MountPodMemLimit:   "2G",
				MountPodCpuRequest: "3",
				MountPodMemRequest: "4G",
			},
			want: testResources,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := parsePodResources(tt.args.MountPodCpuLimit, tt.args.MountPodMemLimit, tt.args.MountPodCpuRequest, tt.args.MountPodMemRequest); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parsePodResources() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getCacheDirVolumes(t *testing.T) {
	cmdWithoutCacheDir := `/bin/mount.juicefs redis://127.0.0.1:6379/0 /jfs/default-imagenet`
	cmdWithCacheDir := `/bin/mount.juicefs redis://127.0.0.1:6379/0 /jfs/default-imagenet -o prefetch=1,cache-dir=/dev/shm/imagenet,cache-size=10240,open-cache=7200,metrics=0.0.0.0:9567`
	cmdWithCacheDir2 := `/bin/mount.juicefs redis://127.0.0.1:6379/0 /jfs/default-imagenet -o cache-dir=/dev/shm/imagenet-0:/dev/shm/imagenet-1,cache-size=10240,metrics=0.0.0.0:9567`

	mp := corev1.MountPropagationBidirectional
	dir := corev1.HostPathDirectory
	volumeMounts := []corev1.VolumeMount{{
		Name:             "jfs-dir",
		MountPath:        config.PodMountBase,
		MountPropagation: &mp,
	}, {
		Name:             "jfs-root-dir",
		MountPath:        "/root/.juicefs",
		MountPropagation: &mp,
	}}

	volumes := []corev1.Volume{{
		Name: "jfs-dir",
		VolumeSource: corev1.VolumeSource{
			HostPath: &corev1.HostPathVolumeSource{
				Path: config.MountPointPath,
				Type: &dir,
			},
		},
	}, {
		Name: "jfs-root-dir",
		VolumeSource: corev1.VolumeSource{
			HostPath: &corev1.HostPathVolumeSource{
				Path: config.JFSConfigPath,
				Type: &dir,
			},
		},
	}}

	cacheVolumes, cacheVolumeMounts := getCacheDirVolumes(cmdWithoutCacheDir)
	volumes = append(volumes, cacheVolumes...)
	volumeMounts = append(volumeMounts, cacheVolumeMounts...)
	if len(volumes) != 2 || len(volumeMounts) != 2 {
		t.Error("getCacheDirVolumes can't work properly")
	}

	cacheVolumes, cacheVolumeMounts = getCacheDirVolumes(cmdWithCacheDir)
	volumes = append(volumes, cacheVolumes...)
	volumeMounts = append(volumeMounts, cacheVolumeMounts...)
	if len(volumes) != 3 || len(volumeMounts) != 3 {
		t.Error("getCacheDirVolumes can't work properly")
	}

	cacheVolumes, cacheVolumeMounts = getCacheDirVolumes(cmdWithCacheDir2)
	volumes = append(volumes, cacheVolumes...)
	volumeMounts = append(volumeMounts, cacheVolumeMounts...)
	if len(volumes) != 5 || len(volumeMounts) != 5 {
		t.Error("getCacheDirVolumes can't work properly")
	}
}

func TestHasRef(t *testing.T) {
	type args struct {
		pod *corev1.Pod
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "test-true",
			args: args{
				pod: &corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:        "test",
						Annotations: map[string]string{"a": "b", "juicefs-aa": "bb"},
					},
				},
			},
			want: true,
		},
		{
			name: "test-false",
			args: args{
				pod: &corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:        "test",
						Annotations: map[string]string{"a": "b"},
					},
				},
			},
			want: false,
		},
		{
			name: "test-null",
			args: args{
				pod: &corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{
						Name:        "test",
						Annotations: nil,
					},
				},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := hasRef(tt.args.pod); got != tt.want {
				t.Errorf("HasRef() = %v, want %v", got, tt.want)
			}
		})
	}
}
