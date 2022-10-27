package interact

import (
        "context"
        "strconv"
        "sync"

        log "github.com/sirupsen/logrus"
        apiv1 "k8s.io/api/core/v1"
        apierrors "k8s.io/apimachinery/pkg/api/errors"
        metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
        "k8s.io/client-go/kubernetes"
        "perf-test/pkg/config"
)

func Create(ctx context.Context, cfg *config.Setup, client *kubernetes.Clientset) error {
        data := make([]byte, cfg.ObjectSizeKB*1000)
        cm := apiv1.ConfigMap{
                TypeMeta: metav1.TypeMeta{
                        Kind:       "ConfigMap",
                        APIVersion: "v1",
                },
                ObjectMeta: metav1.ObjectMeta{
                        Name: "size-test-map",
                },
                BinaryData: map[string][]byte{"data": data},
        }

        wg := &sync.WaitGroup{}
        //operations := make(chan struct{})
        operations := make(chan apiv1.ConfigMap)
        //maps := make(chan apiv1.ConfigMap)
        for i := 0; i < cfg.Parallelism; i++ {
                wg.Add(1)
                go func() {
                        defer wg.Done()
                        for {
                                select {
                                case <-ctx.Done():
                                        return
                                case obj, ok := <-operations:
                                        if !ok {
                                                return
                                        }
                                        _, err := client.CoreV1().ConfigMaps(metav1.NamespaceDefault).Create(ctx, &obj, metav1.CreateOptions{})
                                        if err != nil {
                                                log.Errorf("create config map: %v", err)
                                                if apierrors.IsAlreadyExists(err) {
                                                        client.CoreV1().ConfigMaps(metav1.NamespaceDefault).Delete(ctx, obj.Name, metav1.DeleteOptions{})
                                                }
                                        }
                                }
                        }
                }()
        }

        for i := 0; i < cfg.ObjectCount; i++ {
                cm.ObjectMeta.Name = "size-test-map" + strconv.Itoa(i)
                log.Info(cm.ObjectMeta.Name)
                select {
                case <-ctx.Done():
                        return ctx.Err()
                //case operations <- struct{}{}:
                case operations <- cm:
                }
        }
        close(operations)
        wg.Wait()

        return nil
}
