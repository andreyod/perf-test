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

	mux := &sync.RWMutex{}

	// times map holds times before and after update request send
	times := make(map[string][]time.Time)
	// start watching
	watchObjects(ctx, cfg, wg, times, client, namespace, mux)
	// wait for watcher to be ready
	time.Sleep(1 * time.Minute)

	go metrics.RegisterWatchMetrics()

	// start the threads that will generate events
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
					_, err := client.CoreV1().ConfigMaps(namespace).Patch(ctx, name, types.MergePatchType, []byte(fmt.Sprintf(`{"metadata":{"labels":{"updated":"%d"}}}`, before.Nanosecond())), metav1.PatchOptions{})
					if err != nil {
						log.Errorf("update config map failed: %v", err)
					}
					after := time.Now()
					mux.Lock()
					times[name] = []time.Time{before, after}
					mux.Unlock()
					//log.Infof("update times: %v", times[name])
				}
			}
		}()
	}

	for i := 0; i < cfg.ObjectCount; i++ {
		name := cfg.NamePrefix + strconv.Itoa(i)
		if cfg.UpdateDelay > 0 {
			log.Info("sleep")
			time.Sleep(time.Duration(cfg.UpdateDelay) * time.Second)
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case operations <- name:
		}
	}
	close(operations)
	//time.Sleep(10 * time.Minute)
	wg.Wait()
	return nil
}

func watchObjects(ctx context.Context, cfg *config.Setup, wg *sync.WaitGroup, times map[string][]time.Time, client *kubernetes.Clientset, ns string, mux *sync.RWMutex) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := watch(ctx, cfg, times, client, ns, mux); err != nil && !errors.Is(err, context.DeadlineExceeded) {
			log.WithError(err).Error("failed to interact with the API server")
		}
	}()
}

func watch(ctx context.Context, cfg *config.Setup, times map[string][]time.Time, client *kubernetes.Clientset, ns string, mux *sync.RWMutex) error {
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
			if !ok {
				log.Warning("Watcher channel is closed")
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

			go setMetrics(cfg, obj.GetName(), times, mux, recieved.UnixMilli())

			counter++
			if counter == cfg.ObjectCount {
				log.Infof("%d update events recieved", counter)
				return nil
			}
		}
	}
}

func setMetrics(cfg *config.Setup, name string, times map[string][]time.Time, mux *sync.RWMutex, recieved int64) {
	mux.RLock()
	val, ok := times[name]
	mux.RUnlock()
	if !ok {
		//we got the event before (request returned)/(map values set)
		metrics.TimePoints.WithLabelValues("request-sending").Set(0)
		metrics.TimePoints.WithLabelValues("request-returned").Set(0)
		metrics.TimePoints.WithLabelValues("event-recieved").Set(0)
		log.Info("Event received before request return. Set to 0 latency")
	} else {
		metrics.TimePoints.WithLabelValues("request-sending").Set(float64(val[0].UnixMilli()))
		metrics.TimePoints.WithLabelValues("request-returned").Set(float64(val[1].UnixMilli()))
		metrics.TimePoints.WithLabelValues("event-recieved").Set(float64(recieved))
		log.Infof("name: %s ; latency-ms: %d", name, recieved-val[1].UnixMilli())
	}
}
