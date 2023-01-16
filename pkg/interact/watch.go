package interact

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"sync"
	"time"

	"perf-test/pkg/config"
	"perf-test/pkg/metrics"

	log "github.com/sirupsen/logrus"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
)

func Watch(ctx context.Context, cfg *config.Setup, client *kubernetes.Clientset) error {
	namespace := "test"
	wg := &sync.WaitGroup{}
	// times map holds times before and after update request send
	times := make(map[string][]time.Time)
	// start watching
	watchObjects(ctx, cfg, wg, times, client, namespace)

	//	wg.Add(1)
	//	go func() {
	//		defer wg.Done()
	//		go metrics.RegisterWatchMetrics()
	//	}()
	go metrics.RegisterWatchMetrics()

	// start the threads that will generate events
	//updateForWatching(ctx, im, cfg.Watch, wg)
	operations := make(chan string)
	for i := 0; i < cfg.Parallelism; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				case name, ok := <-operations:
					if !ok {
						return
					}
					before := time.Now()
					//log.Infof("will run update for %s at: %v", name, before)
					//TODO map name-time
					_, err := client.CoreV1().ConfigMaps(namespace).Patch(ctx, name, types.MergePatchType, []byte(fmt.Sprintf(`{"metadata":{"labels":{"updated":"%d"}}}`, before.Nanosecond())), metav1.PatchOptions{})
					if err != nil {
						log.Errorf("update config map failed: %v", err)
					}
					after := time.Now()
					times[name] = []time.Time{before, after}
					//log.Infof("update times: %v", times[name])
				}
			}
		}()
	}

	for i := 0; i < cfg.ObjectCount; i++ {
		name := cfg.NamePrefix + strconv.Itoa(i)
		select {
		case <-ctx.Done():
			return ctx.Err()
		case operations <- name:
		}
	}
	close(operations)
	time.Sleep(10 * time.Minute)
	wg.Wait()
	return nil
}

func watchObjects(ctx context.Context, cfg *config.Setup, wg *sync.WaitGroup, times map[string][]time.Time, client *kubernetes.Clientset, ns string) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := watch(ctx, cfg, times, client, ns); err != nil && !errors.Is(err, context.DeadlineExceeded) {
			log.WithError(err).Error("failed to interact with the API server")
		}
	}()
}

//func watchObjects(ctx context.Context, cfg *config.Setup, wg *sync.WaitGroup, client *kubernetes.Clientset) {
//	for i := 0; i < cfg.Parallelism; i++ {
//		wg.Add(1)
//		go func() {
//			defer wg.Done()
//			if err := watch(ctx, client); err != nil && !errors.Is(err, context.DeadlineExceeded) {
//				log.WithError(err).Error("failed to interact with the API server")
//			}
//		}()
//	}
//}

func watch(ctx context.Context, cfg *config.Setup, times map[string][]time.Time, client *kubernetes.Clientset, ns string) error {
	watcher, err := client.CoreV1().ConfigMaps(ns).Watch(ctx, metav1.ListOptions{})
	if err != nil {
		return err
	}
	counter := 0
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case event, ok := <-watcher.ResultChan():
			recieved := time.Now()
			// i.counterLock.Lock()
			// i.watchCounter++
			// i.counterLock.Unlock()
			if !ok {
				return nil
			}
			switch event.Type {
			case "MODIFIED":
			case "ADDED", "DELETED", "BOOKMARK":
				continue // should not occur
			case "ERROR":
				if err, ok := event.Object.(*metav1.Status); ok && err != nil {
					return fmt.Errorf("watch failed: %w", &apierrors.StatusError{ErrStatus: *err})
				} else {
					return fmt.Errorf("watch failed: %T: %v", event.Object, event.Object)
				}
			}

			obj, ok := event.Object.(metav1.Object)
			if !ok {
				return fmt.Errorf("expected a metav1.Object in watch, got %T", event.Object)
			}
			if val, ok := times[obj.GetName()]; ok {
				metrics.TimePoints.WithLabelValues("request-sending").Set(float64(val[0].UnixMilli()))
				metrics.TimePoints.WithLabelValues("request-returned").Set(float64(val[1].UnixMilli()))
				metrics.TimePoints.WithLabelValues("event-recieved").Set(float64(recieved.UnixMilli()))
			} else {
				log.Errorf("failed to get times for: %s", obj.GetName())
			}

			counter++
			if counter == cfg.ObjectCount {
				log.Infof("%d update events recieved", counter)
				return nil
			}

			// obj, ok := event.Object.(metav1.Object)
			// if !ok {
			// 	return fmt.Errorf("expected a metav1.Object in watch, got %T", event.Object)
			// }
			// i.timestampsLock.RLock()
			// timestamp, ok := i.timestampsByRevision[obj.GetResourceVersion()]
			// i.timestampsLock.RUnlock()
			// if !ok {
			// 	continue
			// }

			// select {
			// case i.metricsChan <- durationDatapoint{method: "watch", duration: recieved.Sub(timestamp)}:
			// case <-ctx.Done():
			// }
		}
	}
}

// func updateForWatching(ctx context.Context, cfg *config.Setup, wg *sync.WaitGroup) {
// 	for i := 0; i < cfg.Parallelism; i++ {
// 		wg.Add(1)
// 		go func() {
// 			defer wg.Done()
// 			for {
// 				select {
// 				case <-ctx.Done():
// 					return
// 				default:
// 				}
// 				if err := update(ctx); err != nil && !errors.Is(err, context.DeadlineExceeded) {
// 					log.WithError(err).Error("failed to interact with the API server")
// 				}
// 			}
// 		}()
// 	}
// }

// func update(ctx context.Context) error {
// 	i.existingLock.Lock()
// 	if len(i.existing) == 0 {
// 		i.existingLock.Unlock()
// 		return i.create(ctx)
// 	}
// 	idx := i.r.Intn(len(i.existing))
// 	key := i.existing[idx]
// 	counter := i.count[key.namespace]
// 	i.existingLock.Unlock()
// 	before := time.Now()
// 	log.Infof("will run update at: %v", before)
// 	rv, err := i.d.update(ctx, key.namespace, key.name, types.MergePatchType, []byte(fmt.Sprintf(`{"metadata":{"labels":{"counter":"%d"}}}`, counter)))
// 	after := time.Now()
// 	duration := after.Sub(before)
// 	i.timestampsLock.Lock()
// 	i.timestampsByRevision[rv] = after
// 	i.timestampsLock.Unlock()
// 	select {
// 	case i.metricsChan <- durationDatapoint{method: "update", duration: duration}:
// 	case <-ctx.Done():
// 	}
// 	return err
// }
