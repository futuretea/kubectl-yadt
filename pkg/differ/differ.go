package differ

import (
	"fmt"

	"github.com/ibuildthecloud/wtfk8s/pkg/printer"
	"github.com/rancher/wrangler/pkg/clients"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

type Differ struct {
	printer      *printer.DiffPrinter
	clients      *clients.Clients
	cache        map[string]*unstructured.Unstructured
	ignoreStatus bool
	ignoreMeta   bool
}

func New(clients *clients.Clients) (*Differ, error) {
	return &Differ{
		printer: printer.NewPrinter(true),
		clients: clients,
		cache:   make(map[string]*unstructured.Unstructured),
	}, nil
}

func (d *Differ) SetIgnoreStatus(ignore bool) {
	d.ignoreStatus = ignore
}

func (d *Differ) SetIgnoreMeta(ignore bool) {
	d.ignoreMeta = ignore
}

func (d *Differ) Print(obj runtime.Object) error {
	unstructuredObj, ok := obj.(*unstructured.Unstructured)
	if !ok {
		data, err := runtime.DefaultUnstructuredConverter.ToUnstructured(obj)
		if err != nil {
			return err
		}
		unstructuredObj = &unstructured.Unstructured{Object: data}
	}

	key := getKey(unstructuredObj)
	oldObj := d.cache[key]

	if oldObj == nil {
		oldObj, _ = d.getCurrentObject(unstructuredObj)
		if oldObj == nil {
			oldObj = newEmptyObject(unstructuredObj)
		}
	}

	d.cache[key] = unstructuredObj.DeepCopy()

	if d.ignoreStatus {
		delete(oldObj.Object, "status")
		delete(unstructuredObj.Object, "status")
	}

	if d.ignoreMeta {
		// Keep only essential metadata fields
		if meta, ok := oldObj.Object["metadata"].(map[string]interface{}); ok {
			cleanMeta := map[string]interface{}{
				"name":      meta["name"],
				"namespace": meta["namespace"],
			}
			oldObj.Object["metadata"] = cleanMeta
		}
		if meta, ok := unstructuredObj.Object["metadata"].(map[string]interface{}); ok {
			cleanMeta := map[string]interface{}{
				"name":      meta["name"],
				"namespace": meta["namespace"],
			}
			unstructuredObj.Object["metadata"] = cleanMeta
		}
	}

	return d.printer.Print(oldObj, unstructuredObj)
}

func (d *Differ) getCurrentObject(obj *unstructured.Unstructured) (*unstructured.Unstructured, error) {
	result, err := d.clients.Dynamic.Get(obj.GroupVersionKind(), obj.GetNamespace(), obj.GetName())
	if err != nil {
		return nil, err
	}

	data, err := runtime.DefaultUnstructuredConverter.ToUnstructured(result)
	if err != nil {
		return nil, err
	}

	return &unstructured.Unstructured{Object: data}, nil
}

func getKey(obj *unstructured.Unstructured) string {
	return fmt.Sprintf("%s/%s/%s/%s",
		obj.GetAPIVersion(),
		obj.GetKind(),
		obj.GetNamespace(),
		obj.GetName())
}

func newEmptyObject(obj *unstructured.Unstructured) *unstructured.Unstructured {
	return &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": obj.GetAPIVersion(),
			"kind":       obj.GetKind(),
			"metadata": map[string]interface{}{
				"name":      obj.GetName(),
				"namespace": obj.GetNamespace(),
			},
		},
	}
}
