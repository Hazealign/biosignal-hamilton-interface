# Biosignal Hamilton Interface

해밀턴 벤틸레이터 G5 기기에서 생체 신호 정보를 받아 NSQ로 보내주는 기기 인터페이스 프로젝트입니다.

**해당 프로젝트를 이해하기 위해서는 해밀턴 벤틸레이터의 RS232 스펙을 이해하는 것이 중요합니다. 스펙 문서가 길지 않기 때문에 이 글을 보기 전에 해당 문서를 정독하는 것을 강하게 권장합니다.**

**Author**: [Haze Lee](https://github.com/Hazealign) / **Last Modify**: February 24, 2017

**ChangeLog**

- **v1.0**: 처음 설계로 구현된 내용을 기반으로 작성되었습니다. [rev. [3db8586](https://github.com/Hazealign/biosignal-hamilton-interface/commit/3db8586440889dd1beccc4590615603e25e57567)]

## PROCESS ARGUMENTS

```go
var Options struct {
	Debug      bool   `short:"d" long:"debug" description:"Enable Debug Mode." optional:"true"`
	Port       string `short:"p" long:"port" description:"Port which connected with Device" required:"true"`
	NsqAddress string `short:"a" long:"address" description:"Address of NSQ Server" required:"true"`
}
```

`-d` 플래그를 통해 디버그 모드를 활성화할 수 있으며, `-p` 플래그를 통해 시리얼 포트를 지정할 수 있으며, `-a` 플래그를 통해 연결할 NSQ 주소를 지정할 수 있습니다.

## HOW WORKS?

1. 프로그램이 시작되면 시리얼 연결이 시작됩니다.
2. 디바이스 ID를 받아오기 위해 시리얼 통신을 1회 주고 받습니다.
3. 이후에는 무한 루프가 돌아갑니다.
   - 디바이스의 Waveform 4개의 값을 받아오기 위한 요청을 보냅니다.(pPatient, pOptional, Volume, Flow)
   - 디바이스가 처리하는데에는 32ms 정도가 걸리기 때문에 36ms 이상을 sleep합니다.
   - 패킷을 읽고 디코딩한 뒤, NSQ에 값을 규격에 맞게 보내줍니다.

## Reference

### packet/predefined_type.go

#### map[int]string: TypeIntString

벤틸레이터 프로토콜에 따라 Identifier를 int, string 형태의 맵으로 맵핑해둔 객체입니다.

#### Response Packet Types

```
RESP_TYPE_RERROR
RESP_TYPE_A
RESP_TYPE_B_FORMAT_1
RESP_TYPE_B_FORMAT_2
RESP_TYPE_B_FORMAT_3
RESP_TYPE_C_34
RESP_TYPE_C_120
```

Go에는 객체지향 패러다임이 존재하지 않기 때문에, 제너릭이나 클래스를 사용할 수 없습니다. 그래서 `ResponsePacket`이라는 단일 구조체를 쓰고 있는데, 장비에서는 여러가지 포맷의 패킷을 줍니다. 편의를 위해 각 타입을 Enum화했습니다.

패킷 포맷의 스펙은 해밀턴 벤틸레이터 RS232 연동 문서의 2.3 ~ 2.5.2를 참고하세요.

### packet/request_packet.go

#### struct: RequestPacket

##### Identifier: byte

장비에 데이터를 요청할 Identifier. 규격은 `predefined_type.go`의 `TypeIntString`를 참고하세요.

#### func: (packet RequestPacket) ToBytes() ([]byte)

구조화된 `RequestPacket` 변수인 `packet`을 패킷으로 인코딩합니다.

#### func: (packet RequestPacket) GetType() (string)

이 Identifier가 어떤 값을 요청하는건지 가져옵니다.

#### func: ParseRequestPacket(raw []byte) (RequestPacket, error)

날 패킷을 `RequestPacket`으로 구조화합니다. 실패 시 에러를 반환합니다.

### packet/response_packet.go

#### struct: ResponsePacket

##### ResponseType: int

패킷의 규격. `predefined_type.go`의 Response Packet Types를 참고하세요.

##### Identifier: byte

요청받은 Identifier. `RequestPacket`의 Identifier와 같습니다.

##### DeviceIdentifier: []byte

장비의 규격 Identifier

##### Values: []byte

장비에서 넘어온 값

##### VentilatorStatus: byte

32번과 120번 Identifier를 위한 벤틸레이터 상태 

##### PPatientLow: byte

32번과 120번 Identifier를 위한 PPatient의 Low Byte

##### PPatientHigh: byte

32번과 120번 Identifier를 위한 PPatient의 High Byte

##### POptionalLow: byte

32번과 120번 Identifier를 위한 POptional의 Low Byte

##### POptionalHigh: byte

32번과 120번 Identifier를 위한 POptional의 High Byte

##### FlowLow: byte

32번과 120번 Identifier를 위한 Flow의 Low Byte

##### FlowHigh: byte

32번과 120번 Identifier를 위한 Flow의 High Byte

##### VolumeLow: byte

32번과 120번 Identifier를 위한 Volume의 Low Byte

##### VolumeHigh: byte

32번과 120번 Identifier를 위한 Volume의 High Byte

##### PCO2Low: byte

32번과 120번 Identifier를 위한 PCO2의 Low Byte

##### PCO2High: byte

32번과 120번 Identifier를 위한 PCO2의 High Byte

#### func: (packet ResponsePacket) ToBytes() ([]byte)

구조화된 `ResponsePacket` 변수인 `packet`을 패킷으로 인코딩합니다.

#### func: ParseResponsePacket(raw []byte) (ResponsePacket, error)

날 패킷을 `ResponsePacket`으로 구조화합니다. 실패 시 에러를 반환합니다.

#### func: ConvertBitWaveform(high byte, low byte) ([]uint8)

High와 Low bit로 쪼개진 두개의 바이트를 2진수 바이너리 배열로 변환합니다.

#### func: BitArrayToInteger(bitArray []uint8) (int)

2진수 바이너리 16 사이즈 배열을 10진수 정수로 변환합니다.

### mq/json_struct.go

#### struct: QueueModel

NSQ에 보내는 데이터 모델, 자세한 규격 설명은 Scheduler 프로젝트의 문서를 참고하세요.

#### func: (d *QueueModel) MarshalJSON() ([]byte, error)

내용을 JSON으로 마샬링합니다. 오류가 발생하면 `error`를 반환합니다.

#### func: SendToNSQ(d QueueModel, str string) (error)

내용을 `str` 채널을 통해 NSQ에 보냅니다. 오류가 발생하면 `error`를 반환합니다.

## Read Also

- [bugst/go-serial](https://github.com/bugst/go-serial)
- [sirupsen/logrus](https://github.com/Sirupsen/logrus)
- [jessevdk/go-flags](https://github.com/jessevdk/go-flags)
- [onsi/ginkgo](https://github.com/onsi/ginkgo)
- [onsi/gomega](https://github.com/onsi/gomega)
- [nsq/go-nsq](https://github.com/nsqio/go-nsq)