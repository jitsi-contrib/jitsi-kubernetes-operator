package controllers

import corev1 "k8s.io/api/core/v1"

func MergeAffinities(affinity *corev1.Affinity, toAdd corev1.Affinity) {
	if affinity.PodAffinity != nil && toAdd.PodAffinity != nil {
		affinity.PodAffinity.PreferredDuringSchedulingIgnoredDuringExecution = append(affinity.PodAffinity.PreferredDuringSchedulingIgnoredDuringExecution, toAdd.PodAffinity.PreferredDuringSchedulingIgnoredDuringExecution...)
		affinity.PodAffinity.RequiredDuringSchedulingIgnoredDuringExecution = append(affinity.PodAffinity.RequiredDuringSchedulingIgnoredDuringExecution, toAdd.PodAffinity.RequiredDuringSchedulingIgnoredDuringExecution...)
	} else if toAdd.PodAffinity != nil {
		affinity.PodAffinity = toAdd.PodAffinity
	}

	if affinity.PodAntiAffinity != nil && toAdd.PodAntiAffinity != nil {
		affinity.PodAntiAffinity.PreferredDuringSchedulingIgnoredDuringExecution = append(affinity.PodAntiAffinity.PreferredDuringSchedulingIgnoredDuringExecution, toAdd.PodAntiAffinity.PreferredDuringSchedulingIgnoredDuringExecution...)
		affinity.PodAntiAffinity.RequiredDuringSchedulingIgnoredDuringExecution = append(affinity.PodAntiAffinity.RequiredDuringSchedulingIgnoredDuringExecution, toAdd.PodAntiAffinity.RequiredDuringSchedulingIgnoredDuringExecution...)
	} else if toAdd.PodAntiAffinity != nil {
		affinity.PodAntiAffinity = toAdd.PodAntiAffinity
	}

	affinity.NodeAffinity = toAdd.NodeAffinity
}
