package ctapi

import (
	"fmt"
	"golang.org/x/sys/windows"
	"os"
	"testing"
)

var tagsToRead = [4]string{"TopMilk_AC01_PV", "TopMilk_AC02_PV", "TopMilk_LS21_PV", "TopMilk_LS07_PV"}

func connectToCitectAPI() (*CtApi, windows.Handle, error) { // Note the addition of error in the return signature
	var dllPath = os.Getenv("CITECT_DLL_PATH")
	api, err := Init("CtApi.dll", dllPath)
	if err != nil {
		return nil, 0, err
	}

	ctapiHandle, err := api.CtOpenRemote("localhost",
		"view",
		"view",
		CT_OPEN_READ_ONLY|CT_OPEN_BATCH|CT_OPEN_EXTENDED|CT_OPEN_RECONNECT)
	fmt.Println(err)
	if err != nil {
		return nil, 0, err
	}

	return api, ctapiHandle, nil
}

func createList(api *CtApi, ctapiHandle windows.Handle) (*CtList, error) {
	list, err := api.NewList(ctapiHandle)
	if err != nil {
		return nil, err
	}
	return list, nil
}

func addTagToList(list *CtList, tag string) (windows.Handle, error) {
	tagHandle, err := list.Add(tag)
	if err != nil {
		fmt.Printf("Failed to add the tag '%s' to the list: %v\n", tag, err)
		return 0, err
	}
	return tagHandle, nil
}

func removeTagFromList(list *CtList, tag string, tagHandle int) bool {
	result, err := list.Delete(windows.Handle(tagHandle))
	if err != nil {
		fmt.Printf("Failed to remove the tag '%s' to the list: %v\n", tag, err)
	}
	return result

}

func TestListAdd(t *testing.T) {
	api, ctapiHandle, err := connectToCitectAPI()
	if err != nil {
		t.Fatalf("Failed to connect to Citect API: %v", err)
	}
	defer api.CtClose(ctapiHandle)

	list, err := createList(api, ctapiHandle)
	if err != nil {
		t.Fatalf("Failed to create list: %v", err)
	}

	tagHandles := make(map[string]int)

	for _, tagName := range tagsToRead {

		tagHandle, err := addTagToList(list, tagName)
		if err != nil {
			t.Fatalf("Failed to add the tag '%s' to the list: %v\n", tagName, err)
		} else {
			fmt.Printf("Tag '%s' to the list: Handle %v\n", tagName, int(tagHandle))
			tagHandles[tagName] = int(tagHandle)
		}
	}

	if err := list.Read(); err != nil {
		t.Fatalf("Failed to read the list: %v", err)
	}

	for tagName, tagHandle := range tagHandles {
		fmt.Println("Reading tag value for", tagName)
		value, err := list.GetFloatValue(windows.Handle(tagHandle))
		if err != nil {
			t.Fatalf("Failed to read the tag '%s' from the list: %v\n", tagName, err)
			return
		} else {
			fmt.Printf("Tag '%s' value: %v\n", tagName, value)
		}
	}

	tagToRemove := "TopMilk_AC01_PV"

	removed := removeTagFromList(list, tagToRemove, tagHandles[tagToRemove])

	// Check the tag was removed successfully, it should return true
	if removed == true {
		fmt.Printf("Tag removed successfully: %s\n", tagToRemove)
	} else {
		t.Fatalf("Failed to remove the tag '%s' from the list\n", tagToRemove)
	}

	// Check if the tag was removed, this should fail returning 0
	result, _ := list.GetFloatValue(windows.Handle(tagHandles[tagToRemove]))
	if result == 0 {
		fmt.Printf("Tag removed successfully: %s\n", tagToRemove)
	}

}
