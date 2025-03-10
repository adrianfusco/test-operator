package ansibletest

import (
	"github.com/openstack-k8s-operators/lib-common/modules/common/env"

	testv1beta1 "github.com/openstack-k8s-operators/test-operator/api/v1beta1"
	util "github.com/openstack-k8s-operators/test-operator/pkg/util"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Job - prepare job to run AnsibleTests tests
func Job(
	instance *testv1beta1.AnsibleTest,
	labels map[string]string,
	jobName string,
	logsPVCName string,
	mountCerts bool,
	envVars map[string]env.Setter,
	workflowOverrideParams map[string]string,
	externalWorkflowCounter int,
	containerImage string,
	privileged bool,
) *batchv1.Job {

	runAsUser := int64(227)
	runAsGroup := int64(227)
	parallelism := int32(1)
	completions := int32(1)

	capabilities := []corev1.Capability{"NET_ADMIN", "NET_RAW"}
	securityContext := util.GetSecurityContext(runAsUser, capabilities, privileged)

	job := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      jobName,
			Namespace: instance.Namespace,
			Labels:    labels,
		},
		Spec: batchv1.JobSpec{
			Parallelism:  &parallelism,
			Completions:  &completions,
			BackoffLimit: instance.Spec.BackoffLimit,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					AutomountServiceAccountToken: &instance.Spec.Privileged,
					RestartPolicy:                corev1.RestartPolicyNever,
					ServiceAccountName:           instance.RbacResourceName(),
					SecurityContext: &corev1.PodSecurityContext{
						RunAsUser:  &runAsUser,
						RunAsGroup: &runAsGroup,
						FSGroup:    &runAsGroup,
					},
					Tolerations:  instance.Spec.Tolerations,
					NodeSelector: instance.Spec.NodeSelector,
					Containers: []corev1.Container{
						{
							Name:            instance.Name,
							Image:           containerImage,
							Args:            []string{},
							Env:             env.MergeEnvs([]corev1.EnvVar{}, envVars),
							VolumeMounts:    GetVolumeMounts(mountCerts, instance, externalWorkflowCounter),
							SecurityContext: &securityContext,
							Resources:       instance.Spec.Resources,
						},
					},
					Volumes: GetVolumes(
						instance,
						logsPVCName,
						mountCerts,
						workflowOverrideParams,
						externalWorkflowCounter,
					),
				},
			},
		},
	}

	if len(instance.Spec.SELinuxLevel) > 0 {
		job.Spec.Template.Spec.SecurityContext.SELinuxOptions = &corev1.SELinuxOptions{
			Level: instance.Spec.SELinuxLevel,
		}
	}

	return job
}
