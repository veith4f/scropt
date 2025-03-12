package controller

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func fqn(script metav1.ObjectMeta) string {
	return script.Namespace + "/" + script.Name
}
