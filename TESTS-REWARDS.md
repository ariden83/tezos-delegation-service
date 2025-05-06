# Tests pour la synchronisation des récompenses Tezos

Nous avons implémenté un module de synchronisation des récompenses (rewards) pour le service de délégation Tezos. Ce document décrit les fonctionnalités implémentées et les tests unitaires associés.

## Implémentation

Le module de synchronisation des récompenses se compose des éléments suivants :

1. **Dans `internal/usecase/sync_rewards.go`** :
   - Structure `SyncRewards` qui gère la logique de synchronisation des récompenses
   - Fonction `NewSyncRewardsFunc` qui crée une nouvelle instance de `SyncRewards`
   - Méthode `Sync` qui synchronise les récompenses depuis l'API TzKT vers la base de données
   - Méthode `saveRewardsBatch` qui sauvegarde un lot de récompenses dans la base de données
   - Méthode `withMonitorer` qui ajoute des capacités de monitoring à la fonction de synchronisation

2. **Dans l'interface de la base de données (`internal/adapter/database/interfaces.go`)** :
   - Méthodes `GetLastSyncedRewardCycle`, `GetActiveDelegators`, `GetBakerForDelegatorAtCycle`, `SaveRewards`, et `SaveLastSyncedRewardCycle`

3. **Dans l'implémentation PostgreSQL (`internal/adapter/database/impl/psql/psql.go`)** :
   - Implémentation des méthodes listées ci-dessus

4. **Dans l'interface TzKT API (`internal/adapter/tzktapi/interfaces.go`)** :
   - Méthodes `GetCurrentCycle` et `FetchRewardsForCycle`

5. **Dans l'implémentation TzKT API (`internal/adapter/tzktapi/impl/api/tzkt.go`)** :
   - Implémentation des méthodes listées ci-dessus

## Tests unitaires

Pour tester correctement cette implémentation, les tests unitaires suivants doivent être créés :

### 1. Test de `NewSyncRewardsFunc`

Ce test vérifie que `NewSyncRewardsFunc` crée correctement une instance de `SyncRewards`.

```go
func Test_NewSyncRewardsFunc(t *testing.T) {
    // Préparer des mocks pour tzktAdapter, dbAdapter, et metricsClient
    tzktAdapter := tzktapimock.New()
    dbAdapter := databasemock.New()
    metricsClient := metricsnoop.New()
    logger := logrus.NewEntry(logrus.New())
    
    // Appeler la fonction
    syncFunc := NewSyncRewardsFunc(tzktAdapter, dbAdapter, metricsClient, logger)
    
    // Vérifier que syncFunc n'est pas nil
    if syncFunc == nil {
        t.Error("Expected non-nil SyncFunc, got nil")
    }
}
```

### 2. Test de la méthode `Sync`

Ce test vérifie que la méthode `Sync` fonctionne correctement dans différents scénarios.

```go
func Test_SyncRewards_Sync(t *testing.T) {
    tests := []struct {
        name    string
        setup   func() (*SyncRewards, context.Context)
        wantErr bool
    }{
        {
            name: "nominal case - up to date",
            setup: func() (*SyncRewards, context.Context) {
                tzktAdapter := tzktapimock.New()
                tzktAdapter.On("GetCurrentCycle", mock.Anything).Return(10, nil)
                
                dbAdapter := databasemock.New()
                dbAdapter.On("GetLastSyncedRewardCycle", mock.Anything).Return(10, nil)
                
                return &SyncRewards{
                    batchSize:      1000,
                    dbAdapter:      dbAdapter,
                    logger:         logrus.NewEntry(logrus.New()),
                    tzktApiAdapter: tzktAdapter,
                }, context.Background()
            },
            wantErr: false,
        },
        // Ajouter d'autres cas de test...
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            uc, ctx := tt.setup()
            err := uc.Sync(ctx)
            if (err != nil) != tt.wantErr {
                t.Errorf("Sync() error = %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}
```

### 3. Test de la méthode `saveRewardsBatch`

Ce test vérifie que la méthode `saveRewardsBatch` fonctionne correctement.

```go
func Test_SyncRewards_saveRewardsBatch(t *testing.T) {
    tests := []struct {
        name    string
        setup   func() (*SyncRewards, context.Context, []model.Reward, int)
        wantErr bool
    }{
        {
            name: "nominal case",
            setup: func() (*SyncRewards, context.Context, []model.Reward, int) {
                dbAdapter := databasemock.New()
                dbAdapter.On("SaveRewards", mock.Anything, mock.Anything).Return(nil)
                
                rewards := []model.Reward{
                    {
                        RecipientAddress: "tz1delegator1",
                        SourceAddress:    "tz1baker1",
                        Cycle:            10,
                        Amount:           5.5,
                        Timestamp:        time.Now().Unix(),
                    },
                }
                
                return &SyncRewards{
                    batchSize: 1000,
                    dbAdapter: dbAdapter,
                    logger:    logrus.NewEntry(logrus.New()),
                }, context.Background(), rewards, 10
            },
            wantErr: false,
        },
        // Ajouter d'autres cas de test...
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            uc, ctx, rewards, cycle := tt.setup()
            err := uc.saveRewardsBatch(ctx, rewards, cycle)
            if (err != nil) != tt.wantErr {
                t.Errorf("saveRewardsBatch() error = %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}
```

### 4. Test de la méthode `withMonitorer`

Ce test vérifie que la méthode `withMonitorer` ajoute correctement le monitoring.

```go
func Test_SyncRewards_withMonitorer(t *testing.T) {
    tests := []struct {
        name            string
        syncRewardsFunc func(context.Context) error
        metricsClient   metrics.Adapter
        wantErr         bool
    }{
        {
            name: "nominal case",
            syncRewardsFunc: func(ctx context.Context) error {
                return nil
            },
            metricsClient: metricsnoop.New(),
            wantErr:      false,
        },
        // Ajouter d'autres cas de test...
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            uc := &SyncRewards{
                logger: logrus.NewEntry(logrus.New()),
            }
            monitoredFunc := uc.withMonitorer(tt.syncRewardsFunc, tt.metricsClient)
            err := monitoredFunc(context.Background())
            if (err != nil) != tt.wantErr {
                t.Errorf("withMonitorer() error = %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}
```

### 5. Tests des méthodes de base de données

Pour les méthodes de base de données implémentées, nous avons ajouté les tests suivants dans `internal/adapter/database/impl/psql/psql_test.go` :
- `Test_psql_GetLastSyncedRewardCycle`
- `Test_psql_GetActiveDelegators`
- `Test_psql_GetBakerForDelegatorAtCycle`
- `Test_psql_SaveRewards`
- `Test_psql_SaveLastSyncedRewardCycle`

### 6. Tests des méthodes TzKT API

Pour les méthodes TzKT API implémentées, nous avons ajouté les tests suivants dans `internal/adapter/tzktapi/impl/api/tzkt_test.go` :
- `Test_Adapter_GetCurrentCycle`
- `Test_Adapter_FetchRewardsForCycle`

## Comment exécuter les tests

Pour exécuter uniquement les tests du package usecase :

```bash
go test ./internal/usecase/
```

Pour exécuter uniquement les tests relatifs aux récompenses :

```bash
go test ./internal/usecase/ -run "Test_SyncRewards_"
go test ./internal/adapter/database/impl/psql/ -run "Test_psql_.*Reward.*"
go test ./internal/adapter/tzktapi/impl/api/ -run "Test_Adapter_.*Cycle.*"
```

## Structure du code

Notre implémentation suit le pattern d'architecture hexagonale (ports et adaptateurs) :
- **Use case** : `SyncRewards` dans `sync_rewards.go`
- **Ports** : interfaces dans `database/interfaces.go` et `tzktapi/interfaces.go`
- **Adaptateurs** : implémentations dans `database/impl/psql/psql.go` et `tzktapi/impl/api/tzkt.go`

Cette séparation permet un testing facile en remplaçant les adaptateurs par des mocks pour tester la logique métier indépendamment des dépendances externes.