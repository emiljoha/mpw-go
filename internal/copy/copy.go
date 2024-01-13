package paste

import (
	"encoding/json"
	"fmt"

	"os"

	"log"

	"github.com/rajveermalviya/go-wayland/wayland/client"
	xdg_shell "github.com/rajveermalviya/go-wayland/wayland/stable/xdg-shell"
)

func WriteText(content string) error {
	display, err := client.Connect("")
	if err != nil {
		log.Fatalf("unable to connect to wayland server: %v", err)
	}
	defer func() {_ = display.Destroy()}()
	printHandlerCalled(display.SetDeleteIdHandler)
	printHandlerCalled(display.SetErrorHandler)	
	registry, err := display.GetRegistry()
	if err != nil {
		log.Fatalf("unable to get global registry object: %v", err)
	}
	defer func() {_ = registry.Destroy()}()
	printHandlerCalled(registry.SetGlobalRemoveHandler)
	var seat *client.Seat
	var dataDeviceManager *client.DataDeviceManager
	var compositor  *client.Compositor
	var shm *client.Shm
	var wmBase *xdg_shell.WmBase
	registry.SetGlobalHandler(func(e client.RegistryGlobalEvent) {
		b, _ := json.Marshal(e)
		fmt.Printf("%T: %s\n", e, string(b))
		switch e.Interface {
		case "wl_seat":
			seat = client.NewSeat(display.Context())
			err := registry.Bind(e.Name, e.Interface, e.Version, seat)
			if err != nil {
				panic(e.Interface +" bind: " + err.Error())
			}			
			printHandlerCalled(seat.SetNameHandler)
		case "wl_data_device_manager":
			dataDeviceManager = client.NewDataDeviceManager(display.Context())
			err := registry.Bind(e.Name, e.Interface, e.Version, dataDeviceManager)
			if err != nil {
				panic("wl_data_device_manager bind: " + err.Error())
			}
		case "wl_compositor":
			compositor = client.NewCompositor(display.Context())
			err := registry.Bind(e.Name, e.Interface, e.Version, compositor)
			if err != nil {
				panic(e.Interface +" bind: " + err.Error())
			}
		case "wl_shm":
			shm = client.NewShm(display.Context())
			err := registry.Bind(e.Name, e.Interface, e.Version, shm)
			if err != nil {
				panic(e.Interface +" bind: " + err.Error())
			}
			printHandlerCalled(shm.SetFormatHandler)
		case "xdg_wm_base":
			wmBase = xdg_shell.NewWmBase(display.Context())
			err := registry.Bind(e.Name, e.Interface, e.Version, wmBase)
			if err != nil {
				panic(e.Interface +" bind: " + err.Error())				
			}
		}
	})
	syncAndWaitForCallback(display)
	defer func() {_ = seat.Release()}()
	defer  func() {_ = dataDeviceManager.Destroy()}()
	defer  func() {_ = compositor.Destroy()}()
	defer  func() {_ = shm.Destroy()}()
	defer  func() {_ = wmBase.Destroy() }()
	wmBase.SetPingHandler(func(e xdg_shell.WmBasePingEvent) {
		b, _ := json.Marshal(e)
		fmt.Printf("%T: %s\n", e, string(b))
		err := wmBase.Pong(e.Serial)
		if err != nil {
			panic(err)
		}
	})
	capabilitiesChan := handlerAsChannel(seat.SetCapabilitiesHandler)		
	syncAndWaitForCallback(display)
	capabilities :=<-capabilitiesChan
	if (capabilities.Capabilities)&(1 << 1) == 0 {
		fmt.Printf(
			"no keyboard capability: %032b\n",
			capabilities.Capabilities,
		)
		panic("no keyboard capability")
	}	
	var dataSources []*client.DataSource
	err = onFocus(wmBase, compositor, seat, display, shm, func() {
		dataSource, err := dataDeviceManager.CreateDataSource()
		if err != nil {
			log.Fatalf("unable to get data source: %v", err)
		}
		printHandlerCalled(dataSource.SetCancelledHandler)
		printHandlerCalled(dataSource.SetTargetHandler)
		printHandlerCalled(dataSource.SetActionHandler)
		printHandlerCalled(dataSource.SetDndDropPerformedHandler)
		printHandlerCalled(dataSource.SetDndFinishedHandler)
		dataSource.SetSendHandler(func(e client.DataSourceSendEvent){
			fmt.Println("SEND!!!!!")
			b, _ := json.Marshal(e)
			fmt.Printf("%T: %s\n", e, string(b))
			f := os.NewFile(uintptr(e.Fd), "file")
			_, err := f.Write([]byte(content))
			if err != nil {
				fmt.Println("send event error: " + err.Error())
			}
			f.Close()
		})
		_ = dataSource.Offer("text/plain")
		_ = dataSource.Offer("TEXT")
		_ = dataSource.Offer("text/plain;charset=utf-8")
		_ = dataSource.Offer("STRING")
		_ = dataSource.Offer("UTF8_STRING")
		syncAndWaitForCallback(display)
		fmt.Println("offers made")
		//fmt.Println("send handler set")
		dataSources = append(dataSources, dataSource)
	})
	if err != nil {
		return err
	}
	fmt.Println("dispatch loop")
	for display.Context().Dispatch() == nil {
	}
	fmt.Println("the end")
	return nil
}

func onFocus(wmBase *xdg_shell.WmBase, compositor *client.Compositor, seat *client.Seat, display *client.Display,
	shm *client.Shm, whenWeHaveFocusCall func()) error {
	fmt.Println("interfaces registered")	
	keyboard, err := seat.GetKeyboard()
	if err != nil {
		return err
	}
	keyboard.SetKeyHandler(func(e client.KeyboardKeyEvent) {
		b, _ := json.Marshal(e)
		fmt.Printf("%T: %s\n", e, string(b))
		whenWeHaveFocusCall()
	})
	printHandlerCalled(keyboard.SetKeymapHandler)
	printHandlerCalled(keyboard.SetLeaveHandler)
	printHandlerCalled(keyboard.SetModifiersHandler)
	printHandlerCalled(keyboard.SetRepeatInfoHandler)
	printHandlerCalled(keyboard.SetEnterHandler)
	//syncAndWaitForCallback(display)
	surface, err := compositor.CreateSurface()
	if err != nil {
		return err
	}
	printHandlerCalled(surface.SetEnterHandler)
	printHandlerCalled(surface.SetLeaveHandler)
	region, err := compositor.CreateRegion()
	if err != nil {
		return err
	}
	surface.SetInputRegion(region)
	xdgSurface, err := wmBase.GetXdgSurface(surface)
	if err != nil {
		return err
	}
	printHandlerCalled(xdgSurface.SetConfigureHandler)
	shellSurface, err := xdgSurface.GetToplevel()
	if err := surface.Commit(); err != nil {
		return err
	}
	printHandlerCalled(shellSurface.SetCloseHandler)
	printHandlerCalled(shellSurface.SetConfigureBoundsHandler)
	printHandlerCalled(shellSurface.SetConfigureHandler)
	printHandlerCalled(shellSurface.SetWmCapabilitiesHandler)
	var width int32 = 100
	var height int32 = 100
	stride := width * 4
	size := stride * height
	tempFile, err := os.CreateTemp("", "*")
	if err != nil {
		return err
	}
	if err = tempFile.Truncate(int64(size)); err != nil {
		return err
	}
	image := make([]byte, size)
	i := 0
	for i < int(size) {
		image[i] = 200
		image[i+1] = 0
		image[i+2] = 100
		image[i+3] = 0
		i = i + 4
	}
	_, err = tempFile.Write(image)
	if err != nil {
		return err
	}
	shmPool, err := shm.CreatePool(int(tempFile.Fd()), int32(size))
	if err != nil {
		return err
	}
	var argb8888 uint32 = 0 // probably...
	buffer, err := shmPool.CreateBuffer(0, width, height, stride, argb8888)
	if err != nil {
		return err
	}
	if err = surface.Attach(buffer, 0, 0); err != nil {
		return err
	}
	if err = surface.Damage(0, 0, width, height); err != nil {
		return err
	}
	if err := surface.Commit(); err != nil {
		return err
	}
	syncAndWaitForCallback(display)
	fmt.Println("popup displayed")
	return nil
}

func printHandlerCalled[H ~func(T), T any](setHandler func(H)) { //nolint	
	setHandler(func(t T) {
		b, _ := json.Marshal(t)
		fmt.Printf("%T: %s\n", t, string(b))
	})
}


func handlerAsChannel[H ~func(T), T any](setHandler func(H)) chan T { //nolint
	tChan := make(chan T)
	setHandler(func(t T) {
		b, _ := json.Marshal(t)
		fmt.Printf("%T: %s\n", t, string(b))
		go func() {tChan <- t}()
	})
	return tChan
}

func syncAndWaitForCallback(display *client.Display) {
	// Get display sync callback
	callback, err := display.Sync()

	if err != nil {
		log.Fatalf("unable to get sync callback: %v", err)
	}
	defer func() {
		if err2 := callback.Destroy(); err2 != nil {
			log.Fatalf("unable to destroy callback: %s", err2.Error())
		}
	}()
	done := false
	callback.SetDoneHandler(func(t client.CallbackDoneEvent) {
		b, _ := json.Marshal(t)
		fmt.Printf("%T: %s\n", t, string(b))
		done = true
	})
	for !done {
		err := display.Context().Dispatch()
		if err != nil {
			fmt.Printf("Dispatch error: %v\n", err)
			return
		}
	}
}
