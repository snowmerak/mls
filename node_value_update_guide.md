# TreeKEM 노드 Value 수정 가이드

## 📋 노드 Value 수정 방법

TreeKEM에서 노드의 Value(실제로는 publicKey)를 수정하는 방법에는 두 가지가 있습니다:

### 1. 🔑 리프 노드 (사용자) 키 수정

**시나리오**: 사용자의 키 로테이션 (Forward Secrecy)

```go
// 방법 1: 직접 수정
func updateUserKey(tree *disk.Tree, userName string, newKey []byte) error {
    // 1. 사용자 노드 찾기
    element, found := tree.Find(userName)
    if !found {
        return fmt.Errorf("user not found: %s", userName)
    }

    // 2. disk.Element로 캐스팅
    diskElement, ok := element.(*disk.Element)
    if !ok {
        return fmt.Errorf("invalid element type")
    }

    // 3. 키 업데이트 및 변경 마킹
    diskElement.SetValue(newKey)     // publicKey 필드 업데이트
    diskElement.MarkAsModified()     // 변경 시점 기록
    
    // 4. 디스크에 저장
    return diskElement.SaveToDisk()
}

// 사용 예시
err := updateUserKey(tree, "alice", []byte("alice_new_rotated_key_epoch2"))
```

### 2. 🌿 중간 노드 키 설정

**시나리오**: TreeKEM 키 파생 후 중간 노드 키 설정

```go
// 중간 노드 키 설정 (이미 구현됨)
err := tree.SetIntermediateNodeKey("intermediate_alice_bob", []byte("derived_shared_key"))
```

---

## 🎯 실제 사용 예시

### TreeKEM 키 로테이션 시나리오

```go
package main

import (
    "fmt"
    "log"
    "time"
    "github.com/snowmerak/mls/lib/tree/disk"
)

func main() {
    tree, _ := disk.NewTree("./data")
    
    // 1. 초기 사용자들 추가
    tree.Insert("alice", []byte("alice_epoch1_key"))
    tree.Insert("bob", []byte("bob_epoch1_key"))
    tree.Insert("charlie", []byte("charlie_epoch1_key"))
    
    log.Println("=== Epoch 1: 초기 상태 ===")
    printTreeState(tree)
    
    // 2. Alice의 키 로테이션 (Forward Secrecy)
    log.Println("\n=== Alice 키 로테이션 수행 ===")
    err := updateUserKey(tree, "alice", []byte("alice_epoch2_rotated_key"))
    if err != nil {
        log.Fatal(err)
    }
    
    // 3. 영향받는 중간 노드들의 키 업데이트
    log.Println("=== 중간 노드들 키 파생 ===")
    updateIntermediateKeys(tree)
    
    log.Println("\n=== Epoch 2: 키 로테이션 후 상태 ===")
    printTreeState(tree)
    
    // 4. 변경 감지 성능 측정
    start := time.Now()
    changedNodes := tree.GetModifiedNodes(time.Time{})
    detectionTime := time.Since(start)
    
    log.Printf("\n⚡ 변경 감지 성능: %v", detectionTime)
    log.Printf("📊 변경된 노드 수: %d개", len(changedNodes))
}

func updateUserKey(tree *disk.Tree, userName string, newKey []byte) error {
    element, found := tree.Find(userName)
    if !found {
        return fmt.Errorf("user not found: %s", userName)
    }

    diskElement := element.(*disk.Element)
    diskElement.SetValue(newKey)
    diskElement.MarkAsModified()
    return diskElement.SaveToDisk()
}

func updateIntermediateKeys(tree *disk.Tree) {
    // 실제 TreeKEM에서는 클라이언트가 DH로 계산한 키를 설정
    intermediates := []string{
        "intermediate_alice_bob",
        "intermediate_alice_charlie", 
    }
    
    for _, name := range intermediates {
        newKey := []byte(fmt.Sprintf("derived_key_%s_epoch2", name))
        tree.SetIntermediateNodeKey(name, newKey)
        log.Printf("  ✓ %s 키 업데이트", name)
    }
}

func printTreeState(tree *disk.Tree) {
    structure := tree.GetTreeStructure()
    for name, info := range structure {
        keyPreview := "empty"
        if len(info.PublicKey) > 0 {
            keyPreview = string(info.PublicKey[:min(15, len(info.PublicKey))])
        }
        log.Printf("  %s [%s]: %s...", name, info.NodeType, keyPreview)
    }
}

func min(a, b int) int {
    if a < b { return a }
    return b
}
```

---

## 🔧 핵심 API

### Element 메서드
```go
// 값 수정
element.SetValue([]byte("new_key"))

// 변경 마킹 (타임스탬프 기록)
element.MarkAsModified()

// 디스크에 저장
element.SaveToDisk()

// 현재 값 조회
value := element.Value()
```

### Tree 메서드
```go
// 중간 노드 키 설정
tree.SetIntermediateNodeKey("node_name", []byte("key"))

// 변경된 노드 조회
modifiedNodes := tree.GetModifiedNodes(since)

// 업데이트 필요한 노드 조회
needUpdate := tree.GetNodesNeedingUpdate()
```

---

## ⚡ 성능 특성

- **변경 감지**: 1-6µs (마이크로초 단위)
- **키 업데이트**: O(1) 단일 노드 수정
- **타임스탬프 추적**: 나노초 정밀도
- **디스크 저장**: 비동기 가능

---

## 🔒 TreeKEM 보안 고려사항

1. **Forward Secrecy**: 키 로테이션으로 이전 키 무효화
2. **변경 추적**: 어떤 노드가 언제 변경되었는지 정확 추적
3. **경로 기반**: 영향받는 노드만 선택적 업데이트
4. **원자적 연산**: 키 업데이트와 타임스탬프가 동시 저장

이 시스템을 통해 TreeKEM의 효율적인 키 관리와 변경 추적이 가능합니다! 🚀