# 설정 파일 레퍼런스

[← 목차로 돌아가기](./USER_GUIDE.md)

---

## servers.csv

서버 목록을 일괄 가져오기/내보내기할 때 사용하는 CSV 파일입니다.

```csv
ip,hostname
192.168.0.1,Router1
192.168.0.2,Switch1
10.0.0.1,CoreSwitch
```

- 첫 번째 행은 헤더(`ip,hostname`)
- 인증 정보는 포함되지 않음 (보안)

---

## commands.txt

실행할 명령어 목록을 가져오기/내보내기할 때 사용하는 텍스트 파일입니다.

```
show version
show ip interface brief
show processes cpu | include CPU
show memory statistics
show running-config
exit
```

- 한 줄에 하나의 명령어
- `exit`는 선택사항 (자동 처리됨)
- 파이프(`|`) 포함 명령어 사용 가능

---

## credentials.json

전역 인증 정보를 저장하는 JSON 파일입니다.

```json
{
  "user": "admin",
  "password": "password"
}
```

---

## config/servers.json (자동 생성)

UI에서 서버별 개별 인증을 설정하면 자동으로 생성됩니다. 인증 정보는 AES-256-GCM으로 암호화됩니다.

```json
[
  {
    "ip": "192.168.0.1",
    "hostname": "Router1",
    "username": "admin",
    "password": "(암호화된 문자열)",
    "enablePassword": "(암호화된 문자열)"
  }
]
```

---

## config/schedules.json (자동 생성)

스케줄 정보가 자동 저장됩니다. 직접 편집하지 않는 것을 권장합니다.

---

## config/smtp.json (자동 생성)

SMTP 설정이 암호화되어 저장됩니다.

```json
{
  "server": "smtp.gmail.com",
  "port": 587,
  "username": "user@gmail.com",
  "password": "(암호화된 문자열)"
}
```

---

## config/encryption.key (자동 생성)

AES-256-GCM 암호화에 사용되는 키 파일입니다. 설치별로 고유하게 생성되며, 삭제 시 저장된 모든 암호화 정보를 읽을 수 없게 됩니다.

---

[← 스케줄링 완전 가이드](./04-scheduling.md) | [다음: FAQ / 트러블슈팅 →](./06-faq.md)
